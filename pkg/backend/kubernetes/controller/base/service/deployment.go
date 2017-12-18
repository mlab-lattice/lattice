package service

import (
	"fmt"
	"reflect"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/controller/base/service/util"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/block"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/golang/glog"
)

func (c *Controller) syncServiceDeployment(service *crv1.Service, nodePool *crv1.NodePool) (*appsv1.Deployment, error) {
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

func (c *Controller) syncExistingDeployment(service *crv1.Service, nodePool *crv1.NodePool, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	// Need a consistent view of our config while generating the deployment spec
	var configCopy *crv1.ConfigSpec
	{
		c.configLock.RLock()
		defer c.configLock.RUnlock()
		configCopy = c.config.DeepCopy()
	}

	name := deploymentName(service)
	labels := deploymentLabels(service)

	desiredSpec, err := c.deploymentSpec(service, name, labels, nodePool, &configCopy.Envoy)
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

func (c *Controller) updateDeploymentSpec(deployment *appsv1.Deployment, spec appsv1.DeploymentSpec) (*appsv1.Deployment, error) {
	if reflect.DeepEqual(deployment.Spec, spec) {
		return deployment, nil
	}

	// Copy so the shared cache isn't mutated
	deployment = deployment.DeepCopy()
	deployment.Spec = spec

	return c.kubeClient.AppsV1().Deployments(deployment.Namespace).Update(deployment)
}

func (c *Controller) createNewDeployment(service *crv1.Service, nodePool *crv1.NodePool) (*appsv1.Deployment, error) {
	deployment, err := c.newDeployment(service, nodePool)
	if err != nil {
		return nil, err
	}

	return c.kubeClient.AppsV1().Deployments(service.Namespace).Create(deployment)
}

func (c *Controller) newDeployment(service *crv1.Service, nodePool *crv1.NodePool) (*appsv1.Deployment, error) {
	var configCopy *crv1.ConfigSpec
	{
		// Need a consistent view of our config while generating the deployment spec
		c.configLock.RLock()
		defer c.configLock.RUnlock()
		configCopy = c.config.DeepCopy()
	}

	name := deploymentName(service)
	labels := deploymentLabels(service)

	spec, err := c.deploymentSpec(service, name, labels, nodePool, &configCopy.Envoy)
	if err != nil {
		return nil, err
	}

	d := &appsv1.Deployment{
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

func (c *Controller) deploymentSpec(service *crv1.Service, name string, deploymentLabels map[string]string, nodePool *crv1.NodePool, envoyConfig *crv1.ConfigEnvoy) (appsv1.DeploymentSpec, error) {
	replicas := service.Spec.NumInstances

	// Create a container for each Component in the Service
	var containers []corev1.Container
	for _, component := range service.Spec.Definition.Components {
		buildArtifacts := service.Spec.ComponentBuildArtifacts[component.Name]
		container := containerFromComponent(component, &buildArtifacts)
		containers = append(containers, container)
	}

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

	spec := &appsv1.DeploymentSpec{
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
				Containers: containers,
				DNSPolicy:  corev1.DNSDefault,
				Affinity: &corev1.Affinity{
					NodeAffinity:    kubeutil.NodePoolNodeAffinity(nodePool),
					PodAntiAffinity: podAntiAffinity,
				},
				Tolerations: []corev1.Toleration{
					kubeutil.NodePoolIDToleration(nodePool),
					kubeutil.NodePoolNamespaceToleration(nodePool),
				},
			},
		},
	}

	spec = c.cloudProvider.TransformServiceDeploymentSpec(spec)
	spec = c.serviceMesh.TransformServiceDeploymentSpec(service, spec)

	return *spec, nil
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
