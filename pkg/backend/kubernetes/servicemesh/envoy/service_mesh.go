package envoy

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/golang/glog"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/lattice/pkg/util/cli"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	netutil "github.com/mlab-lattice/lattice/pkg/util/net"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	annotationKeyAdminPort        = "envoy.servicemesh.lattice.mlab.com/admin-port"
	annotationKeyServiceMeshPorts = "envoy.servicemesh.lattice.mlab.com/service-mesh-ports"
	annotationKeyEgressPorts      = "envoy.servicemesh.lattice.mlab.com/egress-ports"
	annotationKeyIP               = "envoy.servicemesh.lattice.mlab.com/ip"

	deploymentResourcePrefix = "envoy-"

	envoyConfigDirectory           = "/etc/envoy"
	envoyConfigDirectoryVolumeName = deploymentResourcePrefix + "envoyconfig"

	initContainerNamePrepareEnvoy = deploymentResourcePrefix + "prepare-envoy"
	containerNameEnvoy            = deploymentResourcePrefix + "envoy"

	xdsAPIVersion       = "2"
	xdsAPI              = "xds-api"
	labelKeyEnvoyXDSAPI = "envoy.servicemesh.lattice.mlab.com/xds-api"
)

type Options struct {
	PrepareImage      string
	Image             string
	RedirectCIDRBlock net.IPNet
	XDSAPIPort        int32
}

func NewOptions(staticOptions *Options, dynamicConfig *latticev1.ConfigServiceMeshEnvoy) (*Options, error) {
	options := &Options{
		PrepareImage:      dynamicConfig.PrepareImage,
		Image:             dynamicConfig.Image,
		RedirectCIDRBlock: staticOptions.RedirectCIDRBlock,
		XDSAPIPort:        staticOptions.XDSAPIPort,
	}
	return options, nil
}

func NewEnvoyServiceMesh(options *Options) (*DefaultEnvoyServiceMesh, error) {
	leaseManager, err := netutil.NewLeaseManager(options.RedirectCIDRBlock.String())
	if err != nil {
		return nil, err
	}
	// the network IP is reserved for HTTP services, ensure it will not be leased
	err = leaseManager.Blacklist(options.RedirectCIDRBlock.IP.String())
	if err != nil {
		return nil, err
	}
	return &DefaultEnvoyServiceMesh{
		prepareImage:      options.PrepareImage,
		image:             options.Image,
		redirectCIDRBlock: options.RedirectCIDRBlock,
		xdsAPIPort:        options.XDSAPIPort,
		leaseManager:      leaseManager,
	}, nil
}

func Flags() (cli.Flags, *Options) {
	options := &Options{}

	flags := cli.Flags{
		&cli.IPNetFlag{
			Name:     "redirect-cidr-block",
			Required: true,
			Target:   &options.RedirectCIDRBlock,
		},
		&cli.Int32Flag{
			Name:     "xds-api-port",
			Required: true,
			Target:   &options.XDSAPIPort,
		},
	}
	return flags, options
}

type DefaultEnvoyServiceMesh struct {
	prepareImage      string
	image             string
	redirectCIDRBlock net.IPNet
	xdsAPIPort        int32
	leaseManager      netutil.LeaseManager
}

type EnvoyEgressPorts struct {
	HTTP int32 `json:"http"`
	TCP  int32 `json:"tcp"`
}

func (sm *DefaultEnvoyServiceMesh) BootstrapSystemResources(resources *bootstrapper.SystemResources) {
}

func (sm *DefaultEnvoyServiceMesh) ServiceAnnotations(service *latticev1.Service) (map[string]string, error) {
	envoyPorts, err := envoyPorts(service)
	if err != nil {
		return nil, err
	}

	componentPorts, remainingEnvoyPorts, err := assignEnvoyPorts(service, envoyPorts)
	if err != nil {
		return nil, err
	}

	if len(remainingEnvoyPorts) != 3 {
		return nil, fmt.Errorf("expected 3 remaining envoy ports, got %v", len(remainingEnvoyPorts))
	}

	adminPort := remainingEnvoyPorts[0]

	egressPortsJSON, err := json.Marshal(&EnvoyEgressPorts{
		HTTP: remainingEnvoyPorts[1],
		TCP:  remainingEnvoyPorts[2],
	})
	if err != nil {
		return nil, err
	}
	componentPortsJSON, err := json.Marshal(componentPorts)
	if err != nil {
		return nil, err
	}

	annotations := map[string]string{
		annotationKeyAdminPort:        strconv.Itoa(int(adminPort)),
		annotationKeyServiceMeshPorts: string(componentPortsJSON),
		annotationKeyEgressPorts:      string(egressPortsJSON),
	}

	return annotations, nil
}

func (sm *DefaultEnvoyServiceMesh) ServiceAddressAnnotations(
	address *latticev1.Address) (map[string]string, error) {
	ip := address.Annotations[annotationKeyIP]

	annotations := map[string]string{
		annotationKeyIP: ip,
	}

	return annotations, nil
}

func envoyPorts(service *latticev1.Service) ([]int32, error) {
	ports := service.Spec.Definition.ContainerPorts()
	var envoyPortIdx int32 = 10000
	var envoyPorts []int32

	// Need to find len(portSet) + 3 unique ports to use for envoy
	// (two for egress, one for admin, and one per component port for ingress)
	for i := 0; i < len(ports)+3; i++ {

		// Loop up to len(portSet) + 1 times to find an unused port
		// we can use for envoy.
		for j := 0; j < len(ports)+1; j++ {

			// If the current envoyPortIdx is not being used by a component,
			// we'll use it for envoy. Otherwise, on to the next one.
			currPortIdx := envoyPortIdx
			envoyPortIdx++

			if _, ok := ports[currPortIdx]; !ok {
				envoyPorts = append(envoyPorts, currPortIdx)
				break
			}
		}
	}

	if len(envoyPorts) != len(ports)+3 {
		return nil, fmt.Errorf("expected %v envoy ports but got %v", len(ports)+1, len(envoyPorts))
	}

	return envoyPorts, nil
}

func assignEnvoyPorts(service *latticev1.Service, envoyPorts []int32) (map[int32]int32, []int32, error) {
	// Assign an envoy port to each component port, and pop the used envoy port off the slice each time.
	componentPorts := make(map[int32]int32)
	for portNum := range service.Spec.Definition.ContainerPorts() {
		if len(envoyPorts) == 0 {
			return nil, nil, fmt.Errorf("ran out of ports when assigning envoyPorts")
		}

		componentPorts[int32(portNum)] = envoyPorts[0]
		envoyPorts = envoyPorts[1:]
	}

	return componentPorts, envoyPorts, nil
}

func (sm *DefaultEnvoyServiceMesh) TransformServicePodTemplateSpec(
	service *latticev1.Service,
	spec *corev1.PodTemplateSpec,
) (*corev1.PodTemplateSpec, error) {
	prepareEnvoyContainer, envoyContainer, err := sm.envoyContainers(service)
	if err != nil {
		return nil, err
	}

	configVolume := corev1.Volume{
		Name: envoyConfigDirectoryVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	initContainers := []corev1.Container{prepareEnvoyContainer}
	initContainers = append(initContainers, spec.Spec.InitContainers...)

	containers := []corev1.Container{envoyContainer}
	containers = append(containers, spec.Spec.Containers...)

	volumes := []corev1.Volume{configVolume}
	volumes = append(volumes, spec.Spec.Volumes...)

	spec = spec.DeepCopy()

	spec.Spec.InitContainers = initContainers
	spec.Spec.Containers = containers
	spec.Spec.Volumes = volumes
	return spec, nil
}

func (sm *DefaultEnvoyServiceMesh) ServiceMeshPort(service *latticev1.Service, port int32) (int32, error) {
	serviceMeshPorts, err := sm.ServiceMeshPorts(service)
	if err != nil {
		return 0, err
	}

	serviceMeshPort, ok := serviceMeshPorts[port]
	if !ok {
		err := fmt.Errorf(
			"service %v/%v does not have expected port %v",
			service.Namespace,
			service.Name,
			port,
		)
		return 0, err
	}

	return serviceMeshPort, nil
}

func (sm *DefaultEnvoyServiceMesh) ServiceMeshPorts(service *latticev1.Service) (map[int32]int32, error) {
	serviceMeshPortsJSON, ok := service.Annotations[annotationKeyServiceMeshPorts]
	if !ok {
		err := fmt.Errorf(
			"service %v/%v does not have expected annotation %v",
			service.Namespace,
			service.Name,
			serviceMeshPortsJSON,
		)
		return nil, err
	}

	serviceMeshPorts := make(map[int32]int32)
	err := json.Unmarshal([]byte(serviceMeshPortsJSON), &serviceMeshPorts)
	if err != nil {
		return nil, err
	}

	return serviceMeshPorts, nil
}

func (sm *DefaultEnvoyServiceMesh) ServicePort(service *latticev1.Service, port int32) (int32, error) {
	servicePorts, err := sm.ServicePorts(service)
	if err != nil {
		return 0, err
	}

	servicePort, ok := servicePorts[port]
	if !ok {
		err := fmt.Errorf(
			"service %v/%v does not have expected port %v",
			service.Namespace,
			service.Name,
			port,
		)
		return 0, err
	}

	return servicePort, nil
}

func (sm *DefaultEnvoyServiceMesh) ServicePorts(service *latticev1.Service) (map[int32]int32, error) {
	serviceMeshPorts, err := sm.ServiceMeshPorts(service)
	if err != nil {
		return nil, err
	}

	servicePorts := make(map[int32]int32)
	for servicePort, serviceMeshPort := range serviceMeshPorts {
		servicePorts[serviceMeshPort] = servicePort
	}

	return servicePorts, nil
}

func serviceProtocols(service *latticev1.Service) []string {
	protocolSet := make(map[string]interface{})
	for _, componentPort := range service.Spec.Definition.Ports {
		protocolSet[componentPort.Protocol] = nil
	}

	protocols := make([]string, 0, 1)
	for protocol := range protocolSet {
		protocols = append(protocols, protocol)
	}

	return protocols
}

func (sm *DefaultEnvoyServiceMesh) HasServiceIP(address *latticev1.Address) (string, error) {
	annotations, err := sm.ServiceAddressAnnotations(address)
	if err != nil {
		return "", err
	}
	ip := annotations[annotationKeyIP]
	return ip, nil
}

func (sm *DefaultEnvoyServiceMesh) ServiceIP(
	service *latticev1.Service, address *latticev1.Address) (string, map[string]string, error) {
	ip := address.Annotations[annotationKeyIP]

	protocols := serviceProtocols(service)
	if len(protocols) != 1 {
		return "", nil, fmt.Errorf("expected 1 protocol in component ports for service %s, found: %v",
			service.Name, protocols)
	}

	switch protocols[0] {
	case "HTTP":
		netIP := sm.redirectCIDRBlock.IP.String()
		if ip != "" && ip != netIP {
			return "", nil, fmt.Errorf("got IP %s for service %s, expected %s", ip, service.Name, netIP)
		} else {
			ip = netIP
		}
	case "TCP":
		var err error
		ips := make([]string, 0, 1)
		if ip != "" {
			// the lease is already active
			if present, err := sm.leaseManager.IsLeased(ip); err == nil && !present {
				// if the lease manager does not know about the lease, then add it
				// note, this can happen if the address controller dies and restarts
				ips, err = sm.leaseManager.Lease(ip)
			} else {
				ips = append(ips, ip)
			}
		} else {
			// get a new lease from the manager
			ips, err = sm.leaseManager.Lease()
		}
		if err != nil {
			return "", nil, err
		}
		ip = ips[0]
	default:
		return "", nil, fmt.Errorf("expected protocol type HTTP or TCP for service %s, got: %s",
			service.Name, protocols[0])
	}

	annotations, err := sm.ServiceAddressAnnotations(address)
	if err != nil {
		return "", nil, err
	}
	annotations[annotationKeyIP] = ip

	return ip, annotations, nil
}

func (sm *DefaultEnvoyServiceMesh) ReleaseServiceIP(address *latticev1.Address) (map[string]string, error) {
	ip := address.Annotations[annotationKeyIP]

	if ip == "" {
		glog.V(4).Infof("tried to release service IP for %s but found none", address.Name)
		return sm.ServiceAddressAnnotations(address)
	}

	// check if this ip is being managed by
	// XXX <GEB>: race here with call to RemoveLeased, don't believe this is an issue in practice, but may
	//            want to synchronize service mesh methods
	isLeased, err := sm.leaseManager.IsLeased(ip)
	if err != nil {
		return nil, err
	}

	// only remove lease if actually leased (avoids trying to release blacklisted network IP used for
	// HTTP services)
	if isLeased {
		err := sm.leaseManager.RemoveLeased(ip)
		if err != nil {
			return nil, err
		}
	}

	annotations, err := sm.ServiceAddressAnnotations(address)
	if err != nil {
		return nil, err
	}

	// XXX <GEB>: we're setting empty here and passing back to the caller (address controller) to merge into
	//            the Address's annotations. the caller is not expected to diff the current annotations with
	//            the ones we return, so this ensures we overwrite the current lease recorded on the Address's
	//            annotations. revisit and decide whether or not we should go further and actually remove the
	//            key from the set of annotations as well or leave it as metadata (e.g., service mesh was here).
	annotations[annotationKeyIP] = ""

	return annotations, nil
}

func (sm *DefaultEnvoyServiceMesh) IsDeploymentSpecUpdated(
	service *latticev1.Service,
	current, desired, untransformed *appsv1.DeploymentSpec,
) (bool, string, *appsv1.DeploymentSpec) {
	// make sure the init containers are correct
	updated, reason := checkExpectedContainers(current.Template.Spec.InitContainers, desired.Template.Spec.InitContainers, true)
	if !updated {
		return false, reason, nil
	}

	// make sure the containers are correct
	updated, reason = checkExpectedContainers(current.Template.Spec.Containers, desired.Template.Spec.Containers, false)
	if !updated {
		return false, reason, nil
	}

	// make sure the volumes are correct
	updated, reason = checkExpectedVolumes(current.Template.Spec.Volumes, desired.Template.Spec.Volumes)
	if !updated {
		return false, reason, nil
	}

	// get the init containers that are not a part of the serviceMesh
	var initContainers []corev1.Container
	for _, container := range current.Template.Spec.InitContainers {
		if isServiceMeshResource(container.Name) {
			continue
		}

		initContainers = append(initContainers, container)
	}

	// get the containers that are not a part of the serviceMesh
	var containers []corev1.Container
	for _, container := range current.Template.Spec.Containers {
		if isServiceMeshResource(container.Name) {
			continue
		}

		containers = append(containers, container)
	}

	// get the volumes that are not a part of the serviceMesh
	var volumes []corev1.Volume
	for _, volume := range current.Template.Spec.Volumes {
		if isServiceMeshResource(volume.Name) {
			continue
		}

		volumes = append(volumes, volume)
	}

	// make a copy of the desired spec, and set the initContainers, containers, and volumes
	// to be the slices without the service mesh resources
	spec := desired.DeepCopy()
	spec.Template.Spec.InitContainers = initContainers
	spec.Template.Spec.Containers = containers
	spec.Template.Spec.Volumes = volumes

	return true, "", spec
}

func checkExpectedContainers(currentContainers, desiredContainers []corev1.Container, init bool) (bool, string) {
	// Collect all of the expected containers
	desiredEnvoyContainers := map[string]corev1.Container{}
	for _, container := range desiredContainers {
		if !isServiceMeshResource(container.Name) {
			// not a service-mesh init container
			continue
		}

		desiredEnvoyContainers[container.Name] = container
	}

	containerType := ""
	if init {
		containerType = " init"
	}

	// Check to make sure all of the envoy containers exist
	currentEnvoyContainers := make(map[string]interface{})
	for _, container := range currentContainers {
		if !isServiceMeshResource(container.Name) {
			// not a service-mesh init container
			continue
		}

		desiredContainer, ok := desiredEnvoyContainers[container.Name]
		if !ok {
			return false, fmt.Sprintf("has extra envoy%v container %v", containerType, container.Name)
		}

		if !kubeutil.ContainersSemanticallyEqual(&container, &desiredContainer) {
			return false, fmt.Sprintf("has out of date envoy%v container %v", containerType, container.Name)
		}

		currentEnvoyContainers[container.Name] = nil
	}

	// Make sure there aren't extra containers
	numDesiredContainers := len(desiredEnvoyContainers)
	numCurrentContainers := len(currentEnvoyContainers)
	if numDesiredContainers != numCurrentContainers {
		return false, fmt.Sprintf("expected %v envoy%v containers, had %v", numDesiredContainers, containerType, numCurrentContainers)
	}

	return true, ""
}

func checkExpectedVolumes(currentVolumes, desiredVolumes []corev1.Volume) (bool, string) {
	// Collect all of the expected volumes
	desiredEnvoyVolumes := map[string]corev1.Volume{}
	for _, volume := range desiredVolumes {
		if !isServiceMeshResource(volume.Name) {
			// not a service-mesh init volume
			continue
		}

		desiredEnvoyVolumes[volume.Name] = volume
	}

	// Check to make sure all of the volumes exist
	currentEnvoyVolumes := map[string]struct{}{}
	for _, volume := range currentVolumes {
		if !isServiceMeshResource(volume.Name) {
			// not a service-mesh init volume
			continue
		}

		desiredVolume, ok := desiredEnvoyVolumes[volume.Name]
		if !ok {
			return false, fmt.Sprintf("has extra envoy volume %v", volume.Name)
		}

		if !kubeutil.VolumesSemanticallyEqual(&volume, &desiredVolume) {
			return false, fmt.Sprintf("has out of date envoy volume %v", volume.Name)
		}

		currentEnvoyVolumes[volume.Name] = struct{}{}
	}

	numDesiredVolumes := len(desiredEnvoyVolumes)
	numCurrentVolumes := len(currentEnvoyVolumes)
	if numDesiredVolumes != numCurrentVolumes {
		return false, fmt.Sprintf("expected %v envoy volumes, had %v", numDesiredVolumes, numCurrentVolumes)
	}

	return true, ""
}

func isServiceMeshResource(name string) bool {
	parts := strings.Split(name, deploymentResourcePrefix)
	return len(parts) >= 2
}

func (sm *DefaultEnvoyServiceMesh) envoyContainers(service *latticev1.Service) (corev1.Container, corev1.Container, error) {
	adminPort, ok := service.Annotations[annotationKeyAdminPort]
	if !ok {
		err := fmt.Errorf(
			"service %v/%v does not have expected annotation %v",
			service.Namespace,
			service.Name,
			annotationKeyAdminPort,
		)
		return corev1.Container{}, corev1.Container{}, err
	}

	egressPorts, err := sm.EgressPorts(service)
	if err != nil {
		return corev1.Container{}, corev1.Container{}, err
	}

	servicePath, err := service.PathLabel()
	if err != nil {
		return corev1.Container{}, corev1.Container{}, err
	}

	prepareEnvoy := corev1.Container{
		Name:  initContainerNamePrepareEnvoy,
		Image: sm.prepareImage,
		Env: []corev1.EnvVar{
			{
				Name:  "EGRESS_PORT_HTTP",
				Value: strconv.FormatInt(int64(egressPorts.HTTP), 10),
			},
			{
				Name:  "EGRESS_PORT_TCP",
				Value: strconv.FormatInt(int64(egressPorts.TCP), 10),
			},
			{
				Name:  "REDIRECT_EGRESS_CIDR_BLOCK",
				Value: sm.redirectCIDRBlock.String(),
			},
			{
				Name:  "CONFIG_DIR",
				Value: envoyConfigDirectory,
			},
			{
				Name:  "ADMIN_PORT",
				Value: adminPort,
			},
			{
				Name:  "XDS_API_VERSION",
				Value: xdsAPIVersion,
			},
			{
				Name: "XDS_API_HOST",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.hostIP",
					},
				},
			},
			{
				Name:  "XDS_API_PORT",
				Value: fmt.Sprintf("%v", sm.xdsAPIPort),
			},
			// XXX: needed for V2
			{
				Name:  "SERVICE_CLUSTER",
				Value: service.Namespace,
			},
			{
				Name:  "SERVICE_NODE",
				Value: servicePath.ToDomain(),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      envoyConfigDirectoryVolumeName,
				MountPath: envoyConfigDirectory,
			},
		},
		// Need CAP_NET_ADMIN to manipulate iptables
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"NET_ADMIN"},
			},
		},
	}

	var envoyPorts []corev1.ContainerPort
	serviceMeshPorts, err := sm.ServiceMeshPorts(service)
	if err != nil {
		return corev1.Container{}, corev1.Container{}, err
	}

	for portNum := range service.Spec.Definition.ContainerPorts() {
		envoyPort, ok := serviceMeshPorts[portNum]
		if !ok {
			err := fmt.Errorf(
				"service %v/%v does not have expected port %v",
				service.Namespace,
				service.Name,
				portNum,
			)
			return corev1.Container{}, corev1.Container{}, err
		}

		envoyPorts = append(
			envoyPorts,
			corev1.ContainerPort{
				Name:          fmt.Sprintf("%v%v", deploymentResourcePrefix, strconv.Itoa(int(portNum))),
				ContainerPort: envoyPort,
			},
		)
	}

	// XXX: `--service-cluster` and `--service-node` do not seem to have
	//      any effect when running v2 (i.e., they do not set the
	//      service cluster or service node nor do they override whatever
	//      might be set in the config)
	// XXX: adding environment variables to envoy prepare spec to set the
	//      appropriate values in the generated envoy config
	envoy := corev1.Container{
		Name:            containerNameEnvoy,
		Image:           sm.image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/local/bin/envoy"},
		Args: []string{
			"-c", fmt.Sprintf("%v/config.json", envoyConfigDirectory),
			"--service-cluster", service.Namespace,
			"--service-node", servicePath.ToDomain(),
			// by default, the max cluster name size is 60.
			// however, we use the cluster name to encode information, so the names can often be much longer.
			// https://www.envoyproxy.io/docs/envoy/latest/operations/cli#cmdoption-max-obj-name-len
			// FIXME: figure out what this should actually be set to
			"--max-obj-name-len", strconv.Itoa(256),
		},
		Ports: envoyPorts,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      envoyConfigDirectoryVolumeName,
				MountPath: envoyConfigDirectory,
				ReadOnly:  true,
			},
		},
	}

	return prepareEnvoy, envoy, nil
}

func (sm *DefaultEnvoyServiceMesh) EgressPorts(service *latticev1.Service) (*EnvoyEgressPorts, error) {
	egressPortsStr, ok := service.Annotations[annotationKeyEgressPorts]
	if !ok {
		err := fmt.Errorf(
			"service %v/%v does not have expected annotation %v",
			service.Namespace,
			service.Name,
			annotationKeyEgressPorts,
		)
		return nil, err
	}

	var egressPorts EnvoyEgressPorts
	err := json.Unmarshal([]byte(egressPortsStr), &egressPorts)
	if err != nil {
		return nil, err
	}

	return &egressPorts, nil
}
