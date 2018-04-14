package service

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/controller/service/util"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/block"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/golang/glog"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	userResourcePrefix = "lattice-user-"
	ndotsValue         = "15"
	dnsOptionNdots     = "ndots"
)

func (c *Controller) syncServiceDeployment(service *latticev1.Service, nodePool *latticev1.NodePool) (*appsv1.Deployment, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(constants.LabelKeyServiceID, selection.Equals, []string{service.Name})
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

func (c *Controller) syncExistingDeployment(
	service *latticev1.Service,
	nodePool *latticev1.NodePool,
	deployment *appsv1.Deployment,
) (*appsv1.Deployment, error) {
	// Need a consistent view of our config while generating the deployment spec
	c.configLock.RLock()
	defer c.configLock.RUnlock()

	name := deploymentName(service)
	labels := deploymentLabels(service)

	desiredSpec, err := c.deploymentSpec(service, name, labels, nodePool)
	if err != nil {
		return nil, err
	}

	currentSpec := deployment.Spec
	untransformedSpec, err := c.untransformedDeploymentSpec(service, name, labels, nodePool)
	if err != nil {
		return nil, err
	}

	defaultedUntransformedSpec := util.SetDeploymentSpecDefaults(untransformedSpec)
	defaultedDesiredSpec := util.SetDeploymentSpecDefaults(&desiredSpec)

	isUpdated, reason := c.isDeploymentSpecUpdated(service, &currentSpec, defaultedDesiredSpec, defaultedUntransformedSpec)
	if !isUpdated {
		glog.V(4).Infof(
			"Deployment %v for Service %v/%v not up to date: %v",
			deployment.Name,
			service.Namespace,
			service.Name,
			reason,
		)
		return c.updateDeploymentSpec(deployment, desiredSpec)
	}

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

func (c *Controller) createNewDeployment(service *latticev1.Service, nodePool *latticev1.NodePool) (*appsv1.Deployment, error) {
	deployment, err := c.newDeployment(service, nodePool)
	if err != nil {
		return nil, err
	}

	return c.kubeClient.AppsV1().Deployments(service.Namespace).Create(deployment)
}

func (c *Controller) newDeployment(service *latticev1.Service, nodePool *latticev1.NodePool) (*appsv1.Deployment, error) {
	// Need a consistent view of our config while generating the deployment spec
	c.configLock.RLock()
	defer c.configLock.RUnlock()

	name := deploymentName(service)
	labels := deploymentLabels(service)

	spec, err := c.deploymentSpec(service, name, labels, nodePool)
	if err != nil {
		return nil, err
	}

	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*controllerRef(service)},
		},
		Spec: spec,
	}

	return d, nil
}

func deploymentName(service *latticev1.Service) string {
	// TODO(kevinrosendahl): May change this to UUID when a Service can have multiple Deployments (e.g. Blue/Green & Canary)
	return fmt.Sprintf("lattice-service-%s", service.Name)
}

func deploymentLabels(service *latticev1.Service) map[string]string {
	return map[string]string{
		constants.LabelKeyServiceID: service.Name,
	}
}

func (c *Controller) deploymentSpec(
	service *latticev1.Service,
	name string,
	deploymentLabels map[string]string,
	nodePool *latticev1.NodePool,
) (appsv1.DeploymentSpec, error) {
	spec, err := c.untransformedDeploymentSpec(service, name, deploymentLabels, nodePool)
	if err != nil {
		return appsv1.DeploymentSpec{}, err
	}

	// IMPORTANT: the order of these TransformServiceDeploymentSpec and the order of the IsDeploymentSpecUpdated calls in
	// isDeploymentSpecUpdated _must_ be inverses.
	// That is, if we call cloudProvider then serviceMesh here, we _must_ call serviceMesh then cloudProvider
	// in isDeploymentSpecUpdated.
	spec, err = c.serviceMesh.TransformServiceDeploymentSpec(service, spec)
	if err != nil {
		return appsv1.DeploymentSpec{}, err
	}
	spec = c.cloudProvider.TransformServiceDeploymentSpec(service, spec)

	return *spec, nil
}

func (c *Controller) untransformedDeploymentSpec(
	service *latticev1.Service,
	name string,
	deploymentLabels map[string]string,
	nodePool *latticev1.NodePool,
) (*appsv1.DeploymentSpec, error) {
	path, err := service.PathLabel()
	if err != nil {
		// FIXME: in general, if the path label is misformed or missing, the controllers will barf on the whole system.
		// should probably ignore the service and send an error
		return nil, err
	}

	replicas := service.Spec.NumInstances

	// Create a container for each Component in the Service
	var containers []corev1.Container
	for _, component := range service.Spec.Definition.Components() {
		buildArtifacts := service.Spec.ComponentBuildArtifacts[component.Name]
		container, err := containerFromComponent(service, component, &buildArtifacts)
		if err != nil {
			return nil, err
		}

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
		TopologyKey: constants.LabelKeyNodeRoleNodePool,
	}

	// TODO(kevinrosendahl): Make this a PreferredDuringScheduling PodAntiAffinity if the service is running on a shared NodePool
	podAntiAffinity := &corev1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{podAffinityTerm},
	}

	systemID, err := kubeutil.SystemID(service.Namespace)
	if err != nil {
		return nil, err
	}

	baseSearchPath := fmt.Sprintf("%v.%v.local", systemID, c.latticeID)
	dnsSearches := []string{baseSearchPath}

	// If the service is not the root node, we need to append its parent as a search in resolv.conf
	if !path.IsRoot() {
		parentNode, err := path.Parent()
		if err != nil {
			return nil, err
		}

		parentDomain := parentNode.ToDomain()
		dnsSearches = append(dnsSearches, fmt.Sprintf("%v.local.%v", parentDomain, baseSearchPath))
	}

	// as a constant cant be referenced, create a local copy
	ndotsValue := ndotsValue
	dnsConfig := corev1.PodDNSConfig{
		Nameservers: []string{},
		Options: []corev1.PodDNSConfigOption{
			{
				Name:  dnsOptionNdots,
				Value: &ndotsValue,
			},
		},
		Searches: dnsSearches,
	}

	// IMPORTANT: if you change anything in here, you _must_ update isDeploymentSpecUpdated to accommodate it
	spec := &appsv1.DeploymentSpec{
		Replicas: &replicas,
		// FIXME: this is here because envoy currently takes 15 seconds to cycle through XDS API calls
		MinReadySeconds: 15,
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
				DNSConfig:  &dnsConfig,
				Affinity: &corev1.Affinity{
					NodeAffinity:    nodePool.Affinity(),
					PodAntiAffinity: podAntiAffinity,
				},
				Tolerations: []corev1.Toleration{
					nodePool.Toleration(),
				},
			},
		},
	}

	return spec, nil
}

func containerFromComponent(service *latticev1.Service, component *block.Component, buildArtifacts *latticev1.ComponentBuildArtifacts) (corev1.Container, error) {
	path, err := service.PathLabel()
	if err != nil {
		return corev1.Container{}, err
	}

	var ports []corev1.ContainerPort
	for _, port := range component.Ports {
		ports = append(
			ports,
			corev1.ContainerPort{
				Name:          port.Name,
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: int32(port.Port),
			},
		)
	}

	// Sort the env var names so the array order is deterministic
	// so we can more easily check to see if the spec needs
	// to be updated.
	var envVarNames []string
	for name := range component.Exec.Environment {
		envVarNames = append(envVarNames, name)
	}

	sort.Strings(envVarNames)

	var envVars []corev1.EnvVar
	for _, name := range envVarNames {
		envVar := component.Exec.Environment[name]
		if envVar.Value != nil {
			envVars = append(
				envVars,
				corev1.EnvVar{
					Name:  name,
					Value: *envVar.Value,
				},
			)
		} else if envVar.Secret != nil {
			if envVar.Secret.Name != nil {
				envVars = append(
					envVars,
					corev1.EnvVar{
						Name: name,
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: path.ToDomain(),
								},
								Key: *envVar.Secret.Name,
							},
						},
					},
				)
			} else {
				glog.Warning(
					"Component %v for Service %v/%v has environment variable %v which neither has a value or a secret",
					component.Name,
					service.Namespace,
					service.Name,
					name,
				)
			}

			// FIXME: add reference
		}
	}

	probe := deploymentProbe(component.HealthCheck)
	container := corev1.Container{
		Name:            userResourcePrefix + component.Name,
		Image:           buildArtifacts.DockerImageFQN,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         component.Exec.Command,
		Ports:           ports,
		Env:             envVars,
		// TODO(kevinrosendahl): maybe add Resources
		// TODO(kevinrosendahl): add VolumeMounts
		LivenessProbe:  probe,
		ReadinessProbe: probe,
	}
	return container, nil
}

func deploymentProbe(hc *block.ComponentHealthCheck) *corev1.Probe {
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
		// Sort the header names so the array order is deterministic
		// so we can more easily check to see if the spec needs
		// to be updated.
		var headerNames []string
		for name := range hc.HTTP.Headers {
			headerNames = append(headerNames, name)
		}

		sort.Strings(headerNames)

		var headers []corev1.HTTPHeader
		for _, name := range headerNames {
			headers = append(
				headers,
				corev1.HTTPHeader{
					Name:  name,
					Value: hc.HTTP.Headers[name],
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

	// FIXME: should return error here probably
	return nil
}

func (c *Controller) isDeploymentSpecUpdated(
	service *latticev1.Service,
	current, desired, untransformed *appsv1.DeploymentSpec,
) (bool, string) {
	// FIXME: currently Replicas is the only thing we're changing, may want to add other fields to compare as well
	if current.Replicas != desired.Replicas {
		return false, "num replicas is out of date"
	}

	// IMPORTANT: the order of these IsDeploymentSpecUpdated and the order of the TransformServiceDeploymentSpec
	// calls in deploymentSpec _must_ be inverses.
	// That is, if we call serviceMesh then cloudProvider here, we _must_ call cloudProvider then serviceMesh
	// in deploymentSpec.
	// This is done so that IsDeploymentSpecUpdated can return what the spec should look like before it transformed it.
	isUpdated, reason, transformed := c.cloudProvider.IsDeploymentSpecUpdated(service, current, desired, untransformed)
	if !isUpdated {
		return false, reason
	}

	isUpdated, reason, transformed = c.serviceMesh.IsDeploymentSpecUpdated(service, current, transformed, untransformed)
	if !isUpdated {
		return false, reason
	}

	isUpdated = kubeutil.PodTemplateSpecsSemanticallyEqual(&current.Template, &desired.Template)
	if !isUpdated {
		// FIXME: remove when confident this is working correctly
		dmp := diffmatchpatch.New()
		data1, _ := json.MarshalIndent(current.Template, "", "  ")
		data2, _ := json.MarshalIndent(desired.Template, "", "  ")
		diffs := dmp.DiffMain(string(data1), string(data2), true)
		fmt.Printf("diff: %v\n", dmp.DiffPrettyText(diffs))
		return false, "pod template spec is out of date"
	}

	return true, ""
}

type deploymentStatus struct {
	UpdateProcessed bool

	State       deploymentState
	FailureInfo *deploymentStatusFailureInfo

	TotalInstances       int32
	UpdatedInstances     int32
	StaleInstances       int32
	AvailableInstances   int32
	TerminatingInstances int32
}

type deploymentState int

const (
	deploymentStateScaling deploymentState = iota
	deploymentStateStable
	deploymentStateFailed
)

type deploymentStatusFailureInfo struct {
	Reason  string
	Message string
	Time    metav1.Time
}

func (s *deploymentStatus) Failed() (bool, *deploymentStatusFailureInfo) {
	return s.State == deploymentStateFailed, s.FailureInfo
}

func (s *deploymentStatus) Stable() bool {
	return s.State == deploymentStateStable
}

func (c *Controller) getDeploymentStatus(service *latticev1.Service, deployment *appsv1.Deployment) (*deploymentStatus, error) {
	var state deploymentState
	totalInstances := deployment.Status.Replicas
	updatedInstances := deployment.Status.UpdatedReplicas
	availableInstances := deployment.Status.AvailableReplicas
	staleInstances := totalInstances - updatedInstances

	var failureInfo *deploymentStatusFailureInfo
	for _, condition := range deployment.Status.Conditions {
		notProgressing := condition.Type == appsv1.DeploymentProgressing && condition.Status == corev1.ConditionFalse
		if notProgressing && condition.Reason == reasonTimedOut {
			state = deploymentStateFailed
			failureInfo = &deploymentStatusFailureInfo{
				Reason:  condition.Reason,
				Message: condition.Message,
				Time:    condition.LastTransitionTime,
			}
		}
	}

	// Via https://kubernetes.io/docs/concepts/workloads/pods/pod#termination-of-pods,
	// when a pod is Terminating:
	// Pod is removed from endpoints list for service, and are no longer considered part of the set of
	// running pods for replication controllers. Pods that shutdown slowly can continue to serve traffic
	// as load balancers (like the service proxy) remove them from their rotations.
	//
	// That is, when the pod is in Terminating, it has been delivered a SIGTERM but is possibly still running.
	// If, for example, a client has an open connection to the pod, that client can still make requests
	// to the pod. However, at the same time the deployment will not report that this Terminating pod exists.
	// If we were to take the deployment at its word, we could end up saying that this service is stably
	// rolled out to the version specified, even though an old version still exists and could have open
	// connections to it.
	//
	// So if we think that the service is stable, check to see if any pods exist that match are labeled with
	// the service's ID, but have a non-null deletionTimestamp (i.e. they are terminating).
	//
	// TODO: investigate if/how it's possible for pods to get stuck in Terminating, and investigate what
	//       automated processes we can put in place to clean up stuck pods so that deploys don't get stalled
	//       forever
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(constants.LabelKeyServiceID, selection.Equals, []string{service.Name})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	pods, err := c.podLister.Pods(service.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	var terminatingInstances int32
	for _, pod := range pods {
		if pod.DeletionTimestamp != nil {
			terminatingInstances++
		}
	}

	if state != deploymentStateFailed {
		if updatedInstances < totalInstances {
			// The updated pods have not yet all been created
			state = deploymentStateScaling
		} else if totalInstances > updatedInstances {
			// There are extra pods still
			state = deploymentStateScaling
		} else if availableInstances < updatedInstances {
			// there's only updated instances but there aren't enough available instances yet
			state = deploymentStateScaling
		} else if updatedInstances < service.Spec.NumInstances {
			// there only exists UpdatedInstances, and they're all available,
			// but there isn't enough of them yet
			state = deploymentStateScaling
		} else {
			// there are enough available updated instances, and no other instances
			state = deploymentStateStable
		}
	}

	status := &deploymentStatus{
		UpdateProcessed: deployment.Generation <= deployment.Status.ObservedGeneration,

		State:       state,
		FailureInfo: failureInfo,

		TotalInstances:       totalInstances,
		UpdatedInstances:     updatedInstances,
		StaleInstances:       staleInstances,
		AvailableInstances:   availableInstances,
		TerminatingInstances: terminatingInstances,
	}

	return status, nil
}
