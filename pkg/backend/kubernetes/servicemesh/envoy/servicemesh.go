package envoy

import (
	"fmt"
	//"reflect"
	"strconv"
	"strings"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	envoyConfigDirectory           = "/etc/envoy"
	envoyConfigDirectoryVolumeName = "envoyconfig"

	initContainerNamePrepareEnvoy = constants.DeploymentContainerPrefixServiceMesh + "prepare-envoy"
	containerNameEnvoy            = constants.DeploymentContainerPrefixServiceMesh + "envoy"
)

func NewEnvoyServiceMesh(config *crv1.ConfigEnvoy) *DefaultEnvoyServiceMesh {
	return &DefaultEnvoyServiceMesh{
		Config: config,
	}
}

type DefaultEnvoyServiceMesh struct {
	Config *crv1.ConfigEnvoy
}

func (sm *DefaultEnvoyServiceMesh) TransformServiceDeploymentSpec(service *crv1.Service, spec *appsv1.DeploymentSpec) *appsv1.DeploymentSpec {
	prepareEnvoyContainer, envoyContainer := sm.envoyContainers(service)

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

	spec.Template.Spec.InitContainers = initContainers
	spec.Template.Spec.Containers = containers
	spec.Template.Spec.Volumes = volumes
	return spec
}

func (sm *DefaultEnvoyServiceMesh) IsDeploymentSpecUpdated(
	service *crv1.Service,
	current, desired, untransformed *appsv1.DeploymentSpec,
) (bool, string, *appsv1.DeploymentSpec) {
	// Collect all of the envoy init containers
	desiredEnvoyInitContainers := map[string]corev1.Container{}
	for _, container := range desired.Template.Spec.InitContainers {
		parts := strings.Split(container.Name, constants.DeploymentContainerPrefixServiceMesh)
		if len(parts) != 2 {
			// not a service-mesh init container
			continue
		}

		desiredEnvoyInitContainers[container.Name] = container
	}

	return true, "", nil
	// Check to make sure all of the envoy init containers exist, and that there are no extras
	//currentEnvoyInitContainers := map[string]struct{}{}
	//for _, container := range current.Template.Spec.InitContainers {
	//	parts := strings.Split(container.Name, constants.DeploymentContainerPrefixServiceMesh)
	//	if len(parts) != 2 {
	//		// not a service-mesh init container
	//		continue
	//	}
	//
	//	desiredContainer, ok := desiredEnvoyInitContainers[container.Name]
	//	if !ok {
	//		// Has a container that it shouldn't have. current is not Updated
	//		return false, fmt.Sprintf("has extra envoy init container %v", container.Name), nil
	//	}
	//
	//	//if !reflect.DeepEqual(container, desiredContainer) {
	//	//	// container spec is out of date
	//	//}
	//}
}

func (sm *DefaultEnvoyServiceMesh) envoyContainers(service *crv1.Service) (corev1.Container, corev1.Container) {
	prepareEnvoy := corev1.Container{
		Name:  initContainerNamePrepareEnvoy,
		Image: sm.Config.PrepareImage,
		Env: []corev1.EnvVar{
			{
				Name:  "EGRESS_PORT",
				Value: strconv.Itoa(int(service.Spec.EnvoyEgressPort)),
			},
			{
				Name:  "REDIRECT_EGRESS_CIDR_BLOCK",
				Value: sm.Config.RedirectCIDRBlock,
			},
			{
				Name:  "CONFIG_DIR",
				Value: envoyConfigDirectory,
			},
			{
				Name:  "ADMIN_PORT",
				Value: strconv.Itoa(int(service.Spec.EnvoyAdminPort)),
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
				Value: fmt.Sprintf("%v", sm.Config.XDSAPIPort),
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
	for component, ports := range service.Spec.Ports {
		for _, port := range ports {
			envoyPorts = append(
				envoyPorts,
				corev1.ContainerPort{
					Name:          component + "-" + port.Name,
					ContainerPort: port.EnvoyPort,
				},
			)
		}
	}

	envoy := corev1.Container{
		Name:            containerNameEnvoy,
		Image:           sm.Config.Image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/local/bin/envoy"},
		Args: []string{
			"-c",
			fmt.Sprintf("%v/config.json", envoyConfigDirectory),
			"--service-cluster",
			service.Namespace,
			"--service-node",
			service.Spec.Path.ToDomain(false),
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

	return prepareEnvoy, envoy
}
