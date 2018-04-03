package envoy

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	annotationKeyAdminPort        = "envoy.servicemesh.lattice.mlab.com/admin-port"
	annotationKeyServiceMeshPorts = "envoy.servicemesh.lattice.mlab.com/service-mesh-ports"
	annotationKeyEgressPort       = "envoy.servicemesh.lattice.mlab.com/egress-port"

	deploymentResourcePrefix = "lattice-service-mesh-envoy-"

	envoyConfigDirectory           = "/etc/envoy"
	envoyConfigDirectoryVolumeName = deploymentResourcePrefix + "envoyconfig"

	initContainerNamePrepareEnvoy = deploymentResourcePrefix + "prepare-envoy"
	containerNameEnvoy            = deploymentResourcePrefix + "envoy"

	xdsAPI              = "xds-api"
	labelKeyEnvoyXDSAPI = "envoy.servicemesh.lattice.mlab.com/xds-api"
)

type Options struct {
	PrepareImage      string
	Image             string
	RedirectCIDRBlock string
	XDSAPIPort        int32
}

type ServiceMesh interface {
	EgressPort(*latticev1.Service) (int32, error)
	ServiceMeshPort(*latticev1.Service, int32) (int32, error)
	ServiceMeshPorts(*latticev1.Service) (map[int32]int32, error)
	ServicePort(*latticev1.Service, int32) (int32, error)
	ServicePorts(*latticev1.Service) (map[int32]int32, error)
}

func NewEnvoyServiceMesh(options *Options) *DefaultEnvoyServiceMesh {
	return &DefaultEnvoyServiceMesh{
		prepareImage:      options.PrepareImage,
		image:             options.Image,
		redirectCIDRBlock: options.RedirectCIDRBlock,
		xdsAPIPort:        options.XDSAPIPort,
	}
}

type DefaultEnvoyServiceMesh struct {
	prepareImage      string
	image             string
	redirectCIDRBlock string
	xdsAPIPort        int32
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

	if len(remainingEnvoyPorts) != 2 {
		return nil, fmt.Errorf("expected 2 remaining envoy ports, got %v", len(remainingEnvoyPorts))
	}

	adminPort := remainingEnvoyPorts[0]
	egressPort := remainingEnvoyPorts[1]

	componentPortsJSON, err := json.Marshal(componentPorts)
	if err != nil {
		return nil, err
	}

	annotations := map[string]string{
		annotationKeyAdminPort:        strconv.Itoa(int(adminPort)),
		annotationKeyServiceMeshPorts: string(componentPortsJSON),
		annotationKeyEgressPort:       strconv.Itoa(int(egressPort)),
	}

	return annotations, nil
}

func envoyPorts(service *latticev1.Service) ([]int32, error) {
	portSet := map[int32]struct{}{}
	for _, componentPorts := range service.Spec.Ports {
		for _, port := range componentPorts {
			portSet[int32(port.Port)] = struct{}{}
		}
	}

	var envoyPortIdx int32 = 10000
	var envoyPorts []int32

	// Need to find len(portSet) + 2 unique ports to use for envoy
	// (one for egress, one for admin, and one per component port for ingress)
	for i := 0; i <= len(portSet)+1; i++ {

		// Loop up to len(portSet) + 1 times to find an unused port
		// we can use for envoy.
		for j := 0; j <= len(portSet); j++ {

			// If the current envoyPortIdx is not being used by a component,
			// we'll use it for envoy. Otherwise, on to the next one.
			currPortIdx := envoyPortIdx
			envoyPortIdx++

			if _, ok := portSet[currPortIdx]; !ok {
				envoyPorts = append(envoyPorts, currPortIdx)
				break
			}
		}
	}

	if len(envoyPorts) != len(portSet)+2 {
		return nil, fmt.Errorf("expected %v envoy ports but got %v", len(portSet)+1, len(envoyPorts))
	}

	return envoyPorts, nil
}

func assignEnvoyPorts(service *latticev1.Service, envoyPorts []int32) (map[int32]int32, []int32, error) {
	// Assign an envoy port to each component port, and pop the used envoy port off the slice each time.
	componentPorts := map[int32]int32{}
	for _, ports := range service.Spec.Ports {
		for _, port := range ports {
			if len(envoyPorts) == 0 {
				return nil, nil, fmt.Errorf("ran out of ports when assigning envoyPorts")
			}

			componentPorts[port.Port] = envoyPorts[0]
			envoyPorts = envoyPorts[1:]
		}
	}

	return componentPorts, envoyPorts, nil
}

func (sm *DefaultEnvoyServiceMesh) TransformServiceDeploymentSpec(
	service *latticev1.Service,
	spec *appsv1.DeploymentSpec,
) (*appsv1.DeploymentSpec, error) {
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
	initContainers = append(initContainers, spec.Template.Spec.InitContainers...)

	containers := []corev1.Container{envoyContainer}
	containers = append(containers, spec.Template.Spec.Containers...)

	volumes := []corev1.Volume{configVolume}
	volumes = append(volumes, spec.Template.Spec.Volumes...)

	spec = spec.DeepCopy()

	spec.Template.Spec.InitContainers = initContainers
	spec.Template.Spec.Containers = containers
	spec.Template.Spec.Volumes = volumes
	return spec, nil
}

func (sm *DefaultEnvoyServiceMesh) ServicePort(service *latticev1.Service, port int32) (int32, error) {
	servicePorts, err := sm.ServicePorts(service)
	if err != nil {
		return 0, err
	}

	servicePort, ok := servicePorts[port]
	if !ok {
		err := fmt.Errorf(
			"Service %v/%v does not have expected port %v",
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

	servicePorts := map[int32]int32{}
	for servicePort, serviceMeshPort := range serviceMeshPorts {
		servicePorts[serviceMeshPort] = servicePort
	}

	return servicePorts, nil
}

func (sm *DefaultEnvoyServiceMesh) ServiceMeshPort(service *latticev1.Service, port int32) (int32, error) {
	serviceMeshPorts, err := sm.ServiceMeshPorts(service)
	if err != nil {
		return 0, err
	}

	serviceMeshPort, ok := serviceMeshPorts[port]
	if !ok {
		err := fmt.Errorf(
			"Service %v/%v does not have expected port %v",
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
			"Service %v/%v does not have expected annotation %v",
			service.Namespace,
			service.Name,
			serviceMeshPortsJSON,
		)
		return nil, err
	}

	serviceMeshPorts := map[int32]int32{}
	err := json.Unmarshal([]byte(serviceMeshPortsJSON), &serviceMeshPorts)
	if err != nil {
		return nil, err
	}

	return serviceMeshPorts, nil
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
	currentEnvoyContainers := map[string]struct{}{}
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

		currentEnvoyContainers[container.Name] = struct{}{}
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
			"Service %v/%v does not have expected annotation %v",
			service.Namespace,
			service.Name,
			annotationKeyAdminPort,
		)
		return corev1.Container{}, corev1.Container{}, err
	}

	egressPort, err := sm.EgressPort(service)
	if err != nil {
		return corev1.Container{}, corev1.Container{}, err
	}

	prepareEnvoy := corev1.Container{
		Name:  initContainerNamePrepareEnvoy,
		Image: sm.prepareImage,
		Env: []corev1.EnvVar{
			{
				Name:  "EGRESS_PORT",
				Value: strconv.Itoa(int(egressPort)),
			},
			{
				Name:  "REDIRECT_EGRESS_CIDR_BLOCK",
				Value: sm.redirectCIDRBlock,
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

	for component, ports := range service.Spec.Ports {
		for _, port := range ports {
			envoyPort, ok := serviceMeshPorts[port.Port]
			if !ok {
				err := fmt.Errorf(
					"Service %v/%v does not have expected port %v",
					service.Namespace,
					service.Name,
					port,
				)
				return corev1.Container{}, corev1.Container{}, err
			}

			envoyPorts = append(
				envoyPorts,
				corev1.ContainerPort{
					Name:          component + "-" + port.Name,
					ContainerPort: envoyPort,
				},
			)
		}
	}

	servicePath, err := service.PathLabel()
	if err != nil {
		return corev1.Container{}, corev1.Container{}, err
	}

	envoy := corev1.Container{
		Name:            containerNameEnvoy,
		Image:           sm.image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/local/bin/envoy"},
		Args: []string{
			"-c",
			fmt.Sprintf("%v/config.json", envoyConfigDirectory),
			"--service-cluster",
			service.Namespace,
			"--service-node",
			servicePath.ToDomain(),
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

func (sm *DefaultEnvoyServiceMesh) GetEndpointSpec(address *latticev1.ServiceAddress) (*latticev1.EndpointSpec, error) {
	ip, _, err := net.ParseCIDR(sm.redirectCIDRBlock)
	if err != nil {
		return nil, err
	}

	ipStr := ip.String()
	spec := &latticev1.EndpointSpec{
		Path: address.Spec.Path,
		IP:   &ipStr,
	}
	return spec, nil
}

func (sm *DefaultEnvoyServiceMesh) EgressPort(service *latticev1.Service) (int32, error) {
	egressPortStr, ok := service.Annotations[annotationKeyEgressPort]
	if !ok {
		err := fmt.Errorf(
			"Service %v/%v does not have expected annotation %v",
			service.Namespace,
			service.Name,
			annotationKeyEgressPort,
		)
		return 0, err
	}

	egressPort, err := strconv.ParseInt(egressPortStr, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(egressPort), nil
}
