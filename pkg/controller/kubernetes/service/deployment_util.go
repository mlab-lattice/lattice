package service

import (
	"fmt"

	systemdefinitionblock "github.com/mlab-lattice/core/pkg/system/definition/block"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/uuid"
	"strconv"
)

const (
	envoyConfigDirectory           = "/etc/envoy"
	envoyConfigDirectoryVolumeName = "envoyconfig"
)

// getDeployment returns an *extensions.Deployment configured for a given Service
func (sc *ServiceController) getDeployment(svc *crv1.Service) (*extensions.Deployment, error) {
	// Need a consistent view of our config while generating the Job
	sc.configLock.RLock()
	defer sc.configLock.RUnlock()

	name := getDeploymentName(svc)
	labels := map[string]string{
		crv1.ServiceDeploymentLabelKey: svc.Name,
	}

	dSpec, err := sc.getDeploymentSpec(svc)
	if err != nil {
		return nil, err
	}

	d := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(svc, controllerKind)},
		},
		Spec: *dSpec,
	}

	return d, nil
}

func getDeploymentName(svc *crv1.Service) string {
	// TODO: May change this to UUID when a Service can have multiple Deployments (e.g. Blue/Green & Canary)
	return fmt.Sprintf("lattice-service-%s", svc.Name)
}

// getDeploymentSpec returns an *extensions.DeploymentSpec configured for a given Service.
// N.B.: getDeploymentSpec assumes a RLock is held on sc.configLock when called.
func (sc *ServiceController) getDeploymentSpec(svc *crv1.Service) (*extensions.DeploymentSpec, error) {
	containers := []corev1.Container{}
	initContainers := []corev1.Container{}

	// Create a container for each Component in the Service
	for _, component := range svc.Spec.Definition.Components {
		ports := []corev1.ContainerPort{}
		for _, port := range component.Ports {
			ports = append(
				ports,
				corev1.ContainerPort{
					Name:          port.Name,
					ContainerPort: int32(port.Port),
				},
			)
		}

		envs := []corev1.EnvVar{}
		for k, v := range component.Exec.Environment {
			envs = append(
				envs,
				corev1.EnvVar{
					Name:  k,
					Value: v,
				},
			)
		}

		container := corev1.Container{
			Name:    component.Name,
			Image:   svc.Spec.ComponentBuildArtifacts[component.Name].DockerImageFqn,
			Command: component.Exec.Command,
			Ports:   ports,
			Env:     envs,
			// TODO: maybe add Resources
			// TODO: add VolumeMounts
			LivenessProbe: getLivenessProbe(component.HealthCheck),
		}

		if component.Init {
			initContainers = append(initContainers, container)
		} else {
			containers = append(containers, container)
		}
	}

	// Add envoy containers
	envoyConfig := sc.config.Envoy
	initContainers = append(initContainers, corev1.Container{
		// add a UUID to deal with the small chance that a user names their
		// service component the same thing we name our envoy container
		Name:    fmt.Sprintf("lattice-prepare-envoy-%v", uuid.NewUUID()),
		Image:   sc.config.Envoy.PrepareImage,
		Command: []string{"/usr/local/bin/prepare-envoy.sh"},
		Env: []corev1.EnvVar{
			{
				Name:  "ENVOY_EGRESS_PORT",
				Value: strconv.Itoa(int(svc.Spec.EnvoyEgressPort)),
			},
			{
				Name:  "REDIRECT_EGRESS_CIDR_BLOCK",
				Value: envoyConfig.RedirectCidrBlock,
			},
			{
				Name:  "ENVOY_CONFIG_DIR",
				Value: envoyConfigDirectory,
			},
			{
				Name:  "ENVOY_ADMIN_PORT",
				Value: strconv.Itoa(int(svc.Spec.EnvoyAdminPort)),
			},
			{
				Name: "ENVOY_XDS_API_HOST",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.hostIP",
					},
				},
			},
			{
				Name:  "ENVOY_XDS_API_PORT",
				Value: fmt.Sprintf("%v", envoyConfig.XdsApiPort),
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
	})

	envoyPorts := []corev1.ContainerPort{}
	for component, ports := range svc.Spec.Ports {
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
	containers = append(containers, corev1.Container{
		// add a UUID to deal with the small chance that a user names their
		// service component the same thing we name our envoy container
		Name:    fmt.Sprintf("lattice-envoy-%v", uuid.NewUUID()),
		Image:   envoyConfig.Image,
		Command: []string{"/usr/local/bin/envoy"},
		Args: []string{
			"-c",
			fmt.Sprintf("%v/config.json", envoyConfigDirectory),
			"--service-cluster",
			svc.Namespace,
			"--service-node",
			svc.Spec.Path.ToDomain(false),
		},
		Ports: envoyPorts,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      envoyConfigDirectoryVolumeName,
				MountPath: envoyConfigDirectory,
				ReadOnly:  true,
			},
		},
	})

	// Spin up the min instances here, then later let autoscaler scale up.
	// TODO: when doing blue-green deploys, consider looking instead at the current number
	// 		 of replicas in the existing deployment
	replicas := int32(svc.Spec.Definition.Resources.MinInstances)
	deploymentName := getDeploymentName(svc)
	ds := extensions.DeploymentSpec{
		Replicas: &replicas,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: deploymentName,
				Labels: map[string]string{
					crv1.ServiceDeploymentLabelKey: svc.Name,
				},
			},
			Spec: corev1.PodSpec{
				// TODO: add user Volumes
				Volumes: []corev1.Volume{
					// Volume for Envoy config
					{
						Name: envoyConfigDirectoryVolumeName,
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				},
				InitContainers: initContainers,
				Containers:     containers,
				DNSPolicy:      corev1.DNSDefault,
				// FIXME: remove this
				HostAliases: []corev1.HostAlias{
					{
						IP:        "172.16.29.0",
						Hostnames: []string{"private-service.my-system"},
					},
				},
				// TODO: add NodeSelector (for cloud)
				// TODO: add Tolerations (for cloud)
				// TODO: add HostAliases (for local)
			},
		},
	}

	return &ds, nil
}

func getLivenessProbe(hc *systemdefinitionblock.HealthCheck) *corev1.Probe {
	if hc == nil {
		return nil
	}

	if hc.Exec != nil {
		return &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: hc.Exec.Command,
				},
			},
		}
	}

	if hc.Http != nil {
		headers := []corev1.HTTPHeader{}
		for k, v := range hc.Http.Headers {
			headers = append(
				headers,
				corev1.HTTPHeader{
					Name:  k,
					Value: v,
				},
			)
		}

		return &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:        hc.Http.Path,
					Port:        intstr.FromString(hc.Http.Port),
					HTTPHeaders: headers,
				},
			},
		}
	}

	return &corev1.Probe{
		Handler: corev1.Handler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString(hc.Tcp.Port),
			},
		},
	}
}
