package service

import (
	"encoding/json"
	"fmt"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	k8sappsv1 "k8s.io/kubernetes/pkg/apis/apps/v1"

	"github.com/golang/glog"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	ndotsValue     = "15"
	dnsOptionNdots = "ndots"
)

func (c *Controller) syncDeployment(
	service *latticev1.Service,
	nodePool *latticev1.NodePool,
) (*deploymentStatus, error) {
	// FIXME: need to think about implications of rolling deploy between nodes w/ public load balancer
	// probably is just that you need to wait for the address to be updated before syncing an existing deployment
	deployment, err := c.deployment(service)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		// If we need to create a new deployment, we need to wait until the
		// node pool so we can get the right affinity and toleration.
		if nodePool == nil || !nodePool.Stable() {
			return &pendingDeploymentStatus, nil
		}

		return c.createNewDeployment(service, nodePool)
	}

	return c.syncExistingDeployment(service, deployment, nodePool)
}

func (c *Controller) deployment(service *latticev1.Service) (*appsv1.Deployment, error) {
	// First check the cache for the deployment
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.ServiceIDLabelKey, selection.Equals, []string{service.Name})
	if err != nil {
		return nil, err
	}

	selector = selector.Add(*requirement)

	cachedDeployments, err := c.deploymentLister.Deployments(service.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	if len(cachedDeployments) > 1 {
		// This may become valid when doing blue/green deploys
		return nil, fmt.Errorf("found multiple cached deployments for %v", service.Description(c.namespacePrefix))
	}

	if len(cachedDeployments) == 1 {
		return cachedDeployments[0], nil
	}

	// Didn't find the deployment in the cache. This likely means it hasn't been created, but since
	// we can't orphan deployments, we need to do a quorum read first to ensure that the deployment
	// doesn't exist
	deployments, err := c.kubeClient.AppsV1().Deployments(service.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	if len(deployments.Items) > 1 {
		// This may become valid when doing blue/green deploys
		return nil, fmt.Errorf("found multiple deployments for %v", service.Description(c.namespacePrefix))
	}

	if len(deployments.Items) == 1 {
		return &deployments.Items[0], nil
	}

	return nil, nil
}

func (c *Controller) syncExistingDeployment(
	service *latticev1.Service,
	deployment *appsv1.Deployment,
	nodePool *latticev1.NodePool,
) (*deploymentStatus, error) {
	if nodePool == nil {
		return c.getDeploymentStatus(service, deployment)
	}

	currentEpochStable, err := c.currentEpochStable(nodePool)
	if err != nil {
		err := fmt.Errorf(
			"error checking if current epoch for %v node pool is stable: %v",
			service.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	if !currentEpochStable {
		return c.getDeploymentStatus(service, deployment)
	}

	// Need a consistent view of our config while generating the deployment spec
	c.configLock.RLock()
	defer c.configLock.RUnlock()

	name := deploymentName(service)
	deploymentLabels := deploymentLabels(service)
	desiredPodTemplateSpec, err := c.podTemplateSpec(service, name, deploymentLabels, nodePool)
	if err != nil {
		err := fmt.Errorf(
			"error getting desired pod template spec for %v (on %v): %v",
			service.Description(c.namespacePrefix),
			nodePool.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	desiredSpec := c.deploymentSpec(service, deploymentLabels, desiredPodTemplateSpec)
	currentSpec := deployment.Spec

	untransformedPodTemplateSpec, err := c.untransformedPodTemplateSpec(service, name, deploymentLabels, nodePool)
	if err != nil {
		err := fmt.Errorf(
			"error getting untransformed pod template spec for %v (on %v): %v",
			service.Description(c.namespacePrefix),
			nodePool.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	untransformedSpec := c.deploymentSpec(service, deploymentLabels, untransformedPodTemplateSpec)
	defaultedUntransformedSpec := setDeploymentSpecDefaults(&untransformedSpec)
	defaultedDesiredSpec := setDeploymentSpecDefaults(&desiredSpec)

	isUpdated, reason := c.isDeploymentSpecUpdated(service, &currentSpec, defaultedDesiredSpec, defaultedUntransformedSpec)
	if !isUpdated {
		glog.V(4).Infof(
			"deployment %v for %v not up to date: %v",
			deployment.Name,
			service.Description(c.namespacePrefix),
			reason,
		)
		deployment, err = c.updateDeploymentSpec(service, deployment, desiredSpec)
		if err != nil {
			return nil, err
		}
	}

	return c.getDeploymentStatus(service, deployment)
}

func (c *Controller) updateDeploymentSpec(
	service *latticev1.Service,
	deployment *appsv1.Deployment,
	spec appsv1.DeploymentSpec,
) (*appsv1.Deployment, error) {
	if reflect.DeepEqual(deployment.Spec, spec) {
		return deployment, nil
	}

	// Copy so the shared cache isn't mutated
	deployment = deployment.DeepCopy()
	deployment.Spec = spec

	result, err := c.kubeClient.AppsV1().Deployments(deployment.Namespace).Update(deployment)
	if err != nil {
		err := fmt.Errorf(
			"error updating deployment %v for %v: %v",
			deployment.Name,
			service.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	return result, nil
}

func (c *Controller) createNewDeployment(
	service *latticev1.Service,
	nodePool *latticev1.NodePool,
) (*deploymentStatus, error) {
	deployment, err := c.newDeployment(service, nodePool)
	if err != nil {
		return nil, err
	}

	result, err := c.kubeClient.AppsV1().Deployments(service.Namespace).Create(deployment)
	if err != nil {
		err := fmt.Errorf("error creating deployment for %v: %v", service.Description(c.namespacePrefix), err)
		return nil, err
	}

	return c.getDeploymentStatus(service, result)
}

func (c *Controller) newDeployment(service *latticev1.Service, nodePool *latticev1.NodePool) (*appsv1.Deployment, error) {
	// Need a consistent view of our config while generating the deployment spec
	c.configLock.RLock()
	defer c.configLock.RUnlock()

	name := deploymentName(service)
	deploymentLabels := deploymentLabels(service)
	podTemplateSpec, err := c.podTemplateSpec(service, name, deploymentLabels, nodePool)
	if err != nil {
		err := fmt.Errorf(
			"error generating desired pod template spec for %v (on %v): %v",
			service.Description(c.namespacePrefix),
			nodePool.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	spec := c.deploymentSpec(service, deploymentLabels, podTemplateSpec)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          deploymentLabels,
			OwnerReferences: []metav1.OwnerReference{*controllerRef(service)},
		},
		Spec: spec,
	}
	return deployment, nil
}

func deploymentName(service *latticev1.Service) string {
	// TODO(kevinrosendahl): May change this to UUID when a Service can have multiple Deployments (e.g. Blue/Green & Canary)
	return fmt.Sprintf("lattice-service-%s", service.Name)
}

func deploymentLabels(service *latticev1.Service) map[string]string {
	return map[string]string{
		latticev1.ServiceIDLabelKey: service.Name,
	}
}

func (c *Controller) deploymentSpec(
	service *latticev1.Service,
	deploymentLabels map[string]string,
	podTemplateSpec *corev1.PodTemplateSpec,
) appsv1.DeploymentSpec {
	replicas := service.Spec.Definition.NumInstances
	// IMPORTANT: if you change anything in here, you _must_ update isDeploymentSpecUpdated to accommodate it
	return appsv1.DeploymentSpec{
		Replicas: &replicas,
		// FIXME: this is here because envoy currently takes 15 seconds to cycle through XDS API calls
		MinReadySeconds: 15,
		Selector: &metav1.LabelSelector{
			MatchLabels: deploymentLabels,
		},
		Template: *podTemplateSpec,
	}
}

func (c *Controller) podTemplateSpec(
	service *latticev1.Service,
	name string,
	deploymentLabels map[string]string,
	nodePool *latticev1.NodePool,
) (*corev1.PodTemplateSpec, error) {
	podTemplateSpec, err := c.untransformedPodTemplateSpec(service, name, deploymentLabels, nodePool)
	if err != nil {
		return nil, err
	}

	// IMPORTANT: the order of these TransformServicePodTemplateSpec and the order of the IsDeploymentSpecUpdated calls in
	// isDeploymentSpecUpdated _must_ be inverses.
	// That is, if we call cloudProvider then serviceMesh here, we _must_ call serviceMesh then cloudProvider
	// in isDeploymentSpecUpdated.
	podTemplateSpec, err = c.serviceMesh.TransformServicePodTemplateSpec(service, podTemplateSpec)
	if err != nil {
		return nil, err
	}
	podTemplateSpec = c.cloudProvider.TransformPodTemplateSpec(podTemplateSpec)

	return podTemplateSpec, nil
}

func (c *Controller) untransformedPodTemplateSpec(
	service *latticev1.Service,
	name string,
	deploymentLabels map[string]string,
	nodePool *latticev1.NodePool,
) (*corev1.PodTemplateSpec, error) {
	path, err := service.PathLabel()
	if err != nil {
		err := fmt.Errorf("error getting path label for %v: %v", service.Description(c.namespacePrefix), err)
		return nil, err
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
		TopologyKey: latticev1.NodePoolIDLabelKey,
	}

	nodePoolInfo, err := c.nodePoolInfo(service)
	if err != nil {
		return nil, err
	}

	// if the service is running on a per-instance dedicated node pool,
	// then ensure that pods aren't scheduled on the same node.
	// if it's running on a shared node pool, either dedicated or system-level
	// then prefer that pods aren't scheduled on the same node
	podAntiAffinity := &corev1.PodAntiAffinity{}
	switch nodePoolInfo.nodePoolType {
	case latticev1.NodePoolTypeServiceDedicated:
		if nodePoolInfo.perInstance {
			podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = []corev1.PodAffinityTerm{podAffinityTerm}
		} else {
			podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.WeightedPodAffinityTerm{
				{
					Weight:          50,
					PodAffinityTerm: podAffinityTerm,
				},
			}
		}

	case latticev1.NodePoolTypeSystemShared:
		podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.WeightedPodAffinityTerm{
			{
				Weight:          50,
				PodAffinityTerm: podAffinityTerm,
			},
		}

	default:
		err := fmt.Errorf("unrecognized node pool type for %v: %v", service.Description(c.namespacePrefix), nodePoolInfo.nodePoolType)
		return nil, err
	}

	nodePoolEpoch, ok := nodePool.Status.Epochs.CurrentEpoch()
	if !ok {
		return nil, fmt.Errorf("unable to get current epoch for %v: %v", nodePool.Description(c.namespacePrefix), err)
	}

	affinity := &corev1.Affinity{
		NodeAffinity: nodePool.Affinity(nodePoolEpoch),
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight:          50,
					PodAffinityTerm: podAffinityTerm,
				},
			},
		},
	}

	tolerations := []corev1.Toleration{nodePool.Toleration(nodePoolEpoch)}

	return latticev1.PodTemplateSpecForComponent(
		service.Spec.Definition,
		path,
		c.latticeID,
		c.internalDNSDomain,
		c.namespacePrefix,
		service.Namespace,
		service.Name,
		deploymentLabels,
		service.Spec.ContainerBuildArtifacts,
		corev1.RestartPolicyAlways,
		affinity,
		tolerations,
	)
}

func (c *Controller) isDeploymentSpecUpdated(
	service *latticev1.Service,
	current, desired, untransformed *appsv1.DeploymentSpec,
) (bool, string) {
	// NOTE: currently the only thing we change about the top level of the deployment spec (i.e. not in
	// the pod template spec) is replicas. if we change other things we may want to reconsider this
	// comparison strategy.
	if current.Replicas == nil && desired.Replicas != nil {
		return false, "num replicas is out of date"
	}

	if current.Replicas != nil && desired.Replicas == nil {
		return false, "num replicas is out of date"
	}

	if current.Replicas != nil && desired.Replicas != nil && *current.Replicas != *desired.Replicas {
		return false, "num replicas is out of date"
	}

	// FIXME: is this actually true?
	// IMPORTANT: the order of these IsDeploymentSpecUpdated and the order of the TransformServicePodTemplateSpec
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
	deploymentStatePending
)

type deploymentStatusFailureInfo struct {
	Reason string
	Time   metav1.Time
}

var pendingDeploymentStatus = deploymentStatus{
	UpdateProcessed: true,

	State: deploymentStatePending,

	TotalInstances:       0,
	UpdatedInstances:     0,
	StaleInstances:       0,
	AvailableInstances:   0,
	TerminatingInstances: 0,
}

func (s *deploymentStatus) Failed() (bool, *deploymentStatusFailureInfo) {
	return s.State == deploymentStateFailed, s.FailureInfo
}

func (s *deploymentStatus) Stable() bool {
	return s.UpdateProcessed && s.State == deploymentStateStable
}

func (c *Controller) getDeploymentStatus(
	service *latticev1.Service,
	deployment *appsv1.Deployment,
) (*deploymentStatus, error) {
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
				Reason: condition.Reason,
				Time:   condition.LastTransitionTime,
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
	requirement, err := labels.NewRequirement(latticev1.ServiceIDLabelKey, selection.Equals, []string{service.Name})
	if err != nil {
		return nil, fmt.Errorf("error making pod label selector for %v: %v", service.Description(c.namespacePrefix), err)
	}
	selector = selector.Add(*requirement)

	pods, err := c.podLister.Pods(service.Namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("error listing pods for %v: %v", service.Description(c.namespacePrefix), err)
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
		} else if updatedInstances < service.Spec.Definition.NumInstances {
			// there only exists UpdatedInstances, and they're all available,
			// but there isn't enough of them yet
			state = deploymentStateScaling
		} else if terminatingInstances > 0 {
			// There are still pods cleaning up
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

func setDeploymentSpecDefaults(spec *appsv1.DeploymentSpec) *appsv1.DeploymentSpec {
	// Copy so the shared cache isn't mutated
	spec = spec.DeepCopy()

	deployment := &appsv1.Deployment{
		Spec: *spec,
	}
	k8sappsv1.SetObjectDefaults_Deployment(deployment)

	return &deployment.Spec
}
