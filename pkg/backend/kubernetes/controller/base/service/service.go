package service

import (
	"fmt"
	"reflect"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	finalizerName = "service.lattice.mlab.com"

	reasonTimedOut           = "ProgressDeadlineExceeded"
	reasonLoadBalancerFailed = "LoadBalancerFailed"
)

func (c *Controller) syncServiceStatus(
	service *crv1.Service,
	deployment *appsv1.Deployment,
	kubeService *corev1.Service,
	nodePool *crv1.NodePool,
	serviceAddress *crv1.ServiceAddress,
	loadBalancer *crv1.LoadBalancer,
	loadBalancerNeeded bool,
) (*crv1.Service, error) {
	failed := false
	failureReason := ""
	failureMessage := ""
	var failureTime *metav1.Time

	desiredInstances := service.Spec.NumInstances
	updatedInstances := deployment.Status.UpdatedReplicas
	totalInstances := deployment.Status.Replicas
	availableInstances := deployment.Status.AvailableReplicas
	staleInstances := totalInstances - updatedInstances

	for _, condition := range deployment.Status.Conditions {
		notProgressing := condition.Type == appsv1.DeploymentProgressing && condition.Status == corev1.ConditionFalse
		if notProgressing && condition.Reason == reasonTimedOut {
			failed = true
			failureReason = condition.Reason
			failureMessage = condition.Message
			failureTime = &condition.LastTransitionTime
		}
	}

	// First get a basic scaling up/down vs stable state
	// With a little help from: https://github.com/kubernetes/kubernetes/blob/v1.9.0/pkg/kubectl/rollout_status.go#L66-L97
	var state crv1.ServiceState
	if updatedInstances < totalInstances {
		// The updated pods have not yet all been created
		state = crv1.ServiceStateScalingUp
	} else if totalInstances > updatedInstances {
		// There are extra pods still
		state = crv1.ServiceStateScalingDown
	} else if availableInstances < updatedInstances {
		// The update pods have been created but aren't yet available
		state = crv1.ServiceStateScalingUp
	} else {
		state = crv1.ServiceStateStable
	}

	// If the Deployment controller hasn't yet seen the update, it's updating
	if deployment.Generation != deployment.Status.ObservedGeneration {
		state = crv1.ServiceStateUpdating
	} else if state == crv1.ServiceStateStable && desiredInstances != totalInstances {
		// For some reason the Spec is up to date, the deployment is stable, but
		// the deployment does not have the correct number of instances.
		err := fmt.Errorf(
			"Service %v/%v is in state %v but Deployment %v does not have the right amount of instances: expected %v found %v",
			service.Namespace,
			service.Name,
			deployment.Name,
			desiredInstances,
			totalInstances,
		)
		return nil, err
	}

	// If we have any stale instances though, we are updating (which can include scaling)
	// An updating status takes priority over a scaling/stable state
	if staleInstances != 0 {
		state = crv1.ServiceStateUpdating
	}

	// The cloud controller is responsible for creating the Kubernetes Service.
	if kubeService == nil {
		state = crv1.ServiceStateUpdating
	}

	if loadBalancerNeeded {
		switch loadBalancer.Status.State {
		case crv1.LoadBalancerStatePending, crv1.LoadBalancerStateProvisioning:
			state = crv1.ServiceStateUpdating

		case crv1.LoadBalancerStateFailed:
			// Only create new failure info if we didn't fail above
			if !failed {
				now := metav1.Now()
				failed = true
				failureReason = reasonLoadBalancerFailed
				failureMessage = ""
				failureTime = &now
			}
		}
	}

	// But if we have a failure, our updating or scaling has failed
	// A failed status takes priority over an updating status
	var failureInfo *crv1.ServiceFailureInfo
	if failed {
		state = crv1.ServiceStateFailed
		switch failureReason {
		case reasonTimedOut:
			failureInfo = &crv1.ServiceFailureInfo{
				Internal: false,
				Message:  "timed out",
				Time:     *failureTime,
			}

		case reasonLoadBalancerFailed:
			failureInfo = &crv1.ServiceFailureInfo{
				Internal: false,
				Message:  "load balancer failed",
				Time:     *failureTime,
			}

		default:
			failureInfo = &crv1.ServiceFailureInfo{
				Internal: true,
				Message:  fmt.Sprintf("%v: %v", failureReason, failureMessage),
				Time:     *failureTime,
			}
		}
	}

	return c.updateServiceStatus(service, state, updatedInstances, staleInstances, failureInfo)
}

func (c *Controller) updateServiceStatus(
	service *crv1.Service,
	state crv1.ServiceState,
	updatedInstances, staleInstances int32,
	failureInfo *crv1.ServiceFailureInfo,
) (*crv1.Service, error) {
	status := crv1.ServiceStatus{
		State:              state,
		ObservedGeneration: service.Generation,
		UpdatedInstances:   updatedInstances,
		StaleInstances:     staleInstances,
		FailureInfo:        failureInfo,
	}

	if reflect.DeepEqual(service.Status, status) {
		return service, nil
	}

	// Copy the service so the shared cache isn't mutated
	service = service.DeepCopy()
	service.Status = status

	return c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().Services(service.Namespace).UpdateStatus(service)
}

func (c *Controller) syncDeletedService(service *crv1.Service) error {
	lookupFuncs := []func() (interface{}, error){
		func() (interface{}, error) {
			return c.nodePoolLister.NodePools(service.Namespace).Get(service.Name)
		},
		func() (interface{}, error) {
			return c.deploymentLister.Deployments(service.Namespace).Get(service.Name)
		},
		func() (interface{}, error) {
			name := kubeutil.GetKubeServiceNameForService(service.Name)
			return c.kubeServiceLister.Services(service.Namespace).Get(name)
		},
		func() (interface{}, error) {
			return c.serviceAddressLister.ServiceAddresses(service.Namespace).Get(service.Name)
		},
		func() (interface{}, error) {
			return c.loadBalancerLister.LoadBalancers(service.Namespace).Get(service.Name)
		},
	}

	for _, lookupFunc := range lookupFuncs {
		exists, err := resoureceExists(lookupFunc)
		if err != nil {
			return err
		}

		if exists {
			// resource still exists, wait until it is deleted
			return nil
		}
	}

	// All of the children resources have been cleaned up
	_, err := c.removeFinalizer(service)
	return err
}

func resoureceExists(lookupFunc func() (interface{}, error)) (bool, error) {
	_, err := lookupFunc()
	if err == nil {
		// resource still exists, wait until it is deleted
		return true, nil
	}

	if !errors.IsNotFound(err) {
		return false, err
	}

	return false, nil
}

func (c *Controller) addFinalizer(service *crv1.Service) (*crv1.Service, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range service.Finalizers {
		if finalizer == finalizerName {
			glog.V(5).Infof("Endpoint %v has %v finalizer", service.Name, finalizerName)
			return service, nil
		}
	}

	// Add the finalizer to the list and update.
	// If this fails due to a race the Endpoint should get requeued by the controller, so
	// not a big deal.
	service.Finalizers = append(service.Finalizers, finalizerName)
	glog.V(5).Infof("Endpoint %v missing %v finalizer, adding it", service.Name, finalizerName)

	return c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
}

func (c *Controller) removeFinalizer(service *crv1.Service) (*crv1.Service, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range service.Finalizers {
		if finalizer == finalizerName {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return service, nil
	}

	// The finalizer was in the list, so we should remove it.
	service.Finalizers = finalizers
	return c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
}
