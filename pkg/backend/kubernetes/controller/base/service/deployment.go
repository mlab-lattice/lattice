package service

import (
	"fmt"
	"reflect"
	"strconv"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/base/service/util"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/block"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/golang/glog"
	"github.com/satori/go.uuid"
)

const (
	envoyConfigDirectory           = "/etc/envoy"
	envoyConfigDirectoryVolumeName = "envoyconfig"
)

func (c *Controller) syncServiceDeployment(service *crv1.Service, nodePool *crv1.NodePool) (*appsv1beta2.Deployment, error) {
	selector := kubelabels.NewSelector()
	requirement, err := kubelabels.NewRequirement(kubeconstants.LabelKeyServiceID, selection.Equals, []string{service.Name})
	if err != nil {
		return nil, err
	}

	selector = selector.Add(*requirement)
	deployments, err := c.deploymentLister.Deployments(service.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	if len(deployments) > 1 {
		// This may become valid when doing blue/green deploys
		return nil, fmt.Errorf("found multiple deployments for %v/%v", service.Namespace, service.Name)
	}

	if len(deployments) == 0 {
		return c.createNewDeployment(service, nodePool)
	}

	return c.syncExistingDeployment(service, nodePool, deployments[0])
}

func (c *Controller) syncExistingDeployment(service *crv1.Service, nodePool *crv1.NodePool, deployment *appsv1beta2.Deployment) (*appsv1beta2.Deployment, error) {
	// Need a consistent view of our config while generating the deployment spec
	var configCopy *crv1.ConfigSpec
	{
		c.configLock.RLock()
		defer c.configLock.RUnlock()
		configCopy = c.config.DeepCopy()
	}

	name := deploymentName(service)
	labels := deploymentLabels(service)

	desiredSpec, err := deploymentSpec(service, name, labels, nodePool, &configCopy.Envoy)
	if err != nil {
		return nil, err
	}

	desiredPodTemplate := desiredSpec.Template
	podTemplatesSemanticallyEqual := util.PodTemplateSpecsSemanticallyEqual(desiredPodTemplate, deployment.Spec.Template)
	if !podTemplatesSemanticallyEqual {
		glog.V(4).Infof("Deployment %v for Service %v/%v had out of date pod template, updating", deployment.Name, service.Namespace, service.Name)
		return c.updateDeploymentSpec(deployment, desiredSpec)
	}

	if deployment.Spec.Replicas != desiredSpec.Replicas {
		glog.V(4).Infof("Deployment %v for Service %v/%v had out of date number of desired replicas, updating", deployment.Name, service.Namespace, service.Name)
		return c.updateDeploymentSpec(deployment, desiredSpec)
	}

	if deployment.Spec.Strategy != desiredSpec.Strategy {
		glog.V(4).Infof("Deployment %v for Service %v/%v had out of date strategy, updating", deployment.Name, service.Namespace, service.Name)
		return c.updateDeploymentSpec(deployment, desiredSpec)
	}

	if deployment.Spec.MinReadySeconds != desiredSpec.MinReadySeconds {
		glog.V(4).Infof("Deployment %v for Service %v/%v had out of date min ready seconds, updating", deployment.Name, service.Namespace, service.Name)
		return c.updateDeploymentSpec(deployment, desiredSpec)
	}

	if deployment.Spec.Paused != desiredSpec.Paused {
		glog.V(4).Infof("Deployment %v for Service %v/%v had out of paused switch, updating", deployment.Name, service.Namespace, service.Name)
		return c.updateDeploymentSpec(deployment, desiredSpec)
	}

	if deployment.Spec.ProgressDeadlineSeconds != desiredSpec.ProgressDeadlineSeconds {
		glog.V(4).Infof("Deployment %v for Service %v/%v had out of date progress deadline seconds, updating", deployment.Name, service.Namespace, service.Name)
		return c.updateDeploymentSpec(deployment, desiredSpec)
	}

	// It's assumed we won't update the RevisionHistoryLimit or the Selector.
	glog.V(4).Infof("Deployment %v for Service %v/%v was up to date", deployment.Name, service.Namespace, service.Name)
	return deployment, nil
}

func (c *Controller) updateDeploymentSpec(deployment *appsv1beta2.Deployment, spec appsv1beta2.DeploymentSpec) (*appsv1beta2.Deployment, error) {
	if reflect.DeepEqual(deployment.Spec, spec) {
		return deployment, nil
	}

	// Copy so the shared cache isn't mutated
	deployment = deployment.DeepCopy()
	deployment.Spec = spec

	return c.kubeClient.AppsV1beta2().Deployments(deployment.Namespace).Update(deployment)
}

func (c *Controller) createNewDeployment(service *crv1.Service, nodePool *crv1.NodePool) (*appsv1beta2.Deployment, error) {
	deployment, err := c.newDeployment(service, nodePool)
	if err != nil {
		return nil, err
	}

	return c.kubeClient.AppsV1beta2().Deployments(service.Namespace).Create(deployment)
}

func (c *Controller) newDeployment(service *crv1.Service, nodePool *crv1.NodePool) (*appsv1beta2.Deployment, error) {
	var configCopy *crv1.ConfigSpec
	{
		// Need a consistent view of our config while generating the deployment spec
		c.configLock.RLock()
		defer c.configLock.RUnlock()
		configCopy = c.config.DeepCopy()
	}

	name := deploymentName(service)
	labels := deploymentLabels(service)

	spec, err := deploymentSpec(service, name, labels, nodePool, &configCopy.Envoy)
	if err != nil {
		return nil, err
	}

	d := &appsv1beta2.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(service, controllerKind)},
		},
		Spec: spec,
	}

	return d, nil
}

func deploymentName(service *crv1.Service) string {
	// TODO(kevinrosendahl): May change this to UUID when a Service can have multiple Deployments (e.g. Blue/Green & Canary)
	return fmt.Sprintf("lattice-service-%s", service.Name)
}

func deploymentLabels(service *crv1.Service) map[string]string {
	return map[string]string{
		kubeconstants.LabelKeyServiceID: service.Name,
	}
}

func deploymentSpec(service *crv1.Service, name string, deploymentLabels map[string]string, nodePool *crv1.NodePool, envoyConfig *crv1.ConfigEnvoy) (appsv1beta2.DeploymentSpec, error) {
	replicas := service.Spec.NumInstances

	// Create a container for each Component in the Service
	var containers []corev1.Container
	for _, component := range service.Spec.Definition.Components {
		buildArtifacts := service.Spec.ComponentBuildArtifacts[component.Name]
		container := containerFromComponent(component, &buildArtifacts)
		containers = append(containers, container)
	}

	// Add envoy containers
	prepareEnvoyContainer, envoyContainer := envoyContainers(service, envoyConfig)
	initContainers := []corev1.Container{prepareEnvoyContainer}
	containers = append(containers, envoyContainer)

	podAffinityTerm := corev1.PodAffinityTerm{
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: deploymentLabels,
		},
		Namespaces: []string{service.Namespace},

		// This basically tells the pod anti-affinity to only be applied to nodes who all
		// have the same value for that label.
		// Since we also add a RequiredDuringScheduling NodeAffinity for our NodePool,
		// this NodePool's nodes are the only nodes that these pods could be scheduled on,
		// so this TopologyKey doesn't really matter (besides being required).
		TopologyKey: kubeconstants.LabelKeyNodeRoleNodePool,
	}
	// TODO(kevinrosendahl): Make this a PreferredDuringScheduling PodAntiAffinity if the service is running on a shared NodePool
	podAntiAffinity := &corev1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{podAffinityTerm},
	}

	deploymentSpec := appsv1beta2.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: deploymentLabels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: deploymentLabels,
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
				Affinity: &corev1.Affinity{
					NodeAffinity:    kubeutil.NodePoolNodeAffinity(nodePool),
					PodAntiAffinity: podAntiAffinity,
				},
				Tolerations: []corev1.Toleration{
					kubeutil.NodePoolToleration(nodePool),
				},
			},
		},
	}

	return deploymentSpec, nil
}

func containerFromComponent(component *block.Component, buildArtifacts *crv1.ComponentBuildArtifacts) corev1.Container {
	var ports []corev1.ContainerPort
	for _, port := range component.Ports {
		ports = append(
			ports,
			corev1.ContainerPort{
				Name:          port.Name,
				ContainerPort: int32(port.Port),
			},
		)
	}

	var envVars []corev1.EnvVar
	for k, v := range component.Exec.Environment {
		envVars = append(
			envVars,
			corev1.EnvVar{
				Name:  k,
				Value: v,
			},
		)
	}

	return corev1.Container{
		Name:            component.Name,
		Image:           buildArtifacts.DockerImageFQN,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         component.Exec.Command,
		Ports:           ports,
		Env:             envVars,
		// TODO(kevinrosendahl): maybe add Resources
		// TODO(kevinrosendahl): add VolumeMounts
		LivenessProbe: deploymentLivenessProbe(component.HealthCheck),
	}
}

func envoyContainers(service *crv1.Service, config *crv1.ConfigEnvoy) (corev1.Container, corev1.Container) {
	prepareEnvoy := corev1.Container{
		// add a UUID to deal with the small chance that a user names their
		// service component the same thing we name our envoy container
		Name:  fmt.Sprintf("lattice-prepare-envoy-%v", uuid.NewV4().String()),
		Image: config.PrepareImage,
		Env: []corev1.EnvVar{
			{
				Name:  "EGRESS_PORT",
				Value: strconv.Itoa(int(service.Spec.EnvoyEgressPort)),
			},
			{
				Name:  "REDIRECT_EGRESS_CIDR_BLOCK",
				Value: config.RedirectCIDRBlock,
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
				Value: fmt.Sprintf("%v", config.XDSAPIPort),
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
		// add a UUID to deal with the small chance that a user names their
		// service component the same thing we name our envoy container
		Name:            fmt.Sprintf("lattice-envoy-%v", uuid.NewV4().String()),
		Image:           config.Image,
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

func deploymentLivenessProbe(hc *block.ComponentHealthCheck) *corev1.Probe {
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

	if hc.HTTP != nil {
		var headers []corev1.HTTPHeader
		for k, v := range hc.HTTP.Headers {
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
					Path:        hc.HTTP.Path,
					Port:        intstr.FromString(hc.HTTP.Port),
					HTTPHeaders: headers,
				},
			},
		}
	}

	return &corev1.Probe{
		Handler: corev1.Handler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString(hc.TCP.Port),
			},
		},
	}
}
