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
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	netutil "github.com/mlab-lattice/lattice/pkg/util/net"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
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

func (sm *DefaultEnvoyServiceMesh) WorkloadAnnotations(ports map[int32]definitionv1.ContainerPort) (map[string]string, error) {
	envoyPorts, err := envoyPorts(ports)
	if err != nil {
		return nil, err
	}

	componentPorts, remainingEnvoyPorts, err := assignEnvoyPorts(ports, envoyPorts)
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

func (sm *DefaultEnvoyServiceMesh) WorkloadAddressAnnotations(
	address *latticev1.Address) (map[string]string, error) {
	ip := address.Annotations[annotationKeyIP]

	annotations := map[string]string{
		annotationKeyIP: ip,
	}

	return annotations, nil
}

func envoyPorts(ports map[int32]definitionv1.ContainerPort) ([]int32, error) {
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

func assignEnvoyPorts(ports map[int32]definitionv1.ContainerPort, envoyPorts []int32) (map[int32]int32, []int32, error) {
	// Assign an envoy port to each component port, and pop the used envoy port off the slice each time.
	componentPorts := make(map[int32]int32)
	for portNum := range ports {
		if len(envoyPorts) == 0 {
			return nil, nil, fmt.Errorf("ran out of ports when assigning envoyPorts")
		}

		componentPorts[int32(portNum)] = envoyPorts[0]
		envoyPorts = envoyPorts[1:]
	}

	return componentPorts, envoyPorts, nil
}

func (sm *DefaultEnvoyServiceMesh) TransformWorkloadPodTemplateSpec(
	spec *corev1.PodTemplateSpec,
	namespace string,
	componentPath tree.Path,
	annotations map[string]string,
	ports map[int32]definitionv1.ContainerPort,
) (*corev1.PodTemplateSpec, error) {
	prepareEnvoyContainer, envoyContainer, err := sm.envoyContainers(namespace, componentPath, annotations, ports)
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

func (sm *DefaultEnvoyServiceMesh) ServiceMeshPort(annotations map[string]string, port int32) (int32, error) {
	serviceMeshPorts, err := sm.ServiceMeshPorts(annotations)
	if err != nil {
		return 0, err
	}

	serviceMeshPort, ok := serviceMeshPorts[port]
	if !ok {
		err := fmt.Errorf(
			"missing expected port %v",
			port,
		)
		return 0, err
	}

	return serviceMeshPort, nil
}

func (sm *DefaultEnvoyServiceMesh) ServiceMeshPorts(annotations map[string]string) (map[int32]int32, error) {
	serviceMeshPortsJSON, ok := annotations[annotationKeyServiceMeshPorts]
	if !ok {
		err := fmt.Errorf(
			"missing expected annotation %v",
			annotationKeyServiceMeshPorts,
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

func (sm *DefaultEnvoyServiceMesh) WorkloadPort(annotations map[string]string, port int32) (int32, error) {
	servicePorts, err := sm.WorkloadPorts(annotations)
	if err != nil {
		return 0, err
	}

	servicePort, ok := servicePorts[port]
	if !ok {
		err := fmt.Errorf(
			"missing expected port %v",
			port,
		)
		return 0, err
	}

	return servicePort, nil
}

func (sm *DefaultEnvoyServiceMesh) WorkloadPorts(annotations map[string]string) (map[int32]int32, error) {
	serviceMeshPorts, err := sm.ServiceMeshPorts(annotations)
	if err != nil {
		return nil, err
	}

	servicePorts := make(map[int32]int32)
	for servicePort, serviceMeshPort := range serviceMeshPorts {
		servicePorts[serviceMeshPort] = servicePort
	}

	return servicePorts, nil
}

func workloadProtocols(ports map[int32]definitionv1.ContainerPort) []string {
	protocolSet := make(map[string]interface{})
	for _, componentPort := range ports {
		protocolSet[componentPort.Protocol] = nil
	}

	protocols := make([]string, 0, 1)
	for protocol := range protocolSet {
		protocols = append(protocols, protocol)
	}

	return protocols
}

func (sm *DefaultEnvoyServiceMesh) HasWorkloadIP(address *latticev1.Address) (string, error) {
	annotations, err := sm.WorkloadAddressAnnotations(address)
	if err != nil {
		return "", err
	}
	ip := annotations[annotationKeyIP]
	return ip, nil
}

func (sm *DefaultEnvoyServiceMesh) WorkloadIP(
	address *latticev1.Address,
	workloadPorts map[int32]definitionv1.ContainerPort,
) (string, map[string]string, error) {
	ip := address.Annotations[annotationKeyIP]

	protocols := workloadProtocols(workloadPorts)
	if len(protocols) != 1 {
		return "", nil, fmt.Errorf("expected 1 protocol in component ports, found: %v", protocols)
	}

	switch protocols[0] {
	case "HTTP":
		netIP := sm.redirectCIDRBlock.IP.String()
		if ip != "" && ip != netIP {
			return "", nil, fmt.Errorf("got IP %s, expected %s", ip, netIP)
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
		return "", nil, fmt.Errorf("expected protocol type HTTP or TCP got: %s", protocols[0])
	}

	annotations, err := sm.WorkloadAddressAnnotations(address)
	if err != nil {
		return "", nil, err
	}
	annotations[annotationKeyIP] = ip

	return ip, annotations, nil
}

func (sm *DefaultEnvoyServiceMesh) ReleaseWorkloadIP(address *latticev1.Address) (map[string]string, error) {
	ip, _ := address.Annotations[annotationKeyIP]

	if ip == "" {
		glog.V(4).Infof("tried to release service IP for %s but found none", address.Name)
		return sm.WorkloadAddressAnnotations(address)
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

	annotations, err := sm.WorkloadAddressAnnotations(address)
	if err != nil {
		return nil, err
	}
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

func (sm *DefaultEnvoyServiceMesh) envoyContainers(
	namespace string,
	componentPath tree.Path,
	annotations map[string]string,
	ports map[int32]definitionv1.ContainerPort,
) (corev1.Container, corev1.Container, error) {
	adminPort, ok := annotations[annotationKeyAdminPort]
	if !ok {
		err := fmt.Errorf(
			"does not have expected annotation %v",
			annotationKeyAdminPort,
		)
		return corev1.Container{}, corev1.Container{}, err
	}

	egressPorts, err := sm.egressPorts(annotations)
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
				Value: namespace,
			},
			{
				Name:  "SERVICE_NODE",
				Value: componentPath.ToDomain(),
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
	serviceMeshPorts, err := sm.ServiceMeshPorts(annotations)
	if err != nil {
		return corev1.Container{}, corev1.Container{}, err
	}

	for portNum := range ports {
		envoyPort, ok := serviceMeshPorts[portNum]
		if !ok {
			err := fmt.Errorf(
				"does not have expected port %v",
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
			"--service-cluster", namespace,
			"--service-node", componentPath.ToDomain(),
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

func (sm *DefaultEnvoyServiceMesh) egressPorts(annotations map[string]string) (*EnvoyEgressPorts, error) {
	egressPortsStr, ok := annotations[annotationKeyEgressPorts]
	if !ok {
		err := fmt.Errorf(
			"missing expected annotation %v",
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
