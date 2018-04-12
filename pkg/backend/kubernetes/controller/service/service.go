package service

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	finalizerName = "service.lattice.mlab.com"

	reasonTimedOut           = "ProgressDeadlineExceeded"
	reasonLoadBalancerFailed = "LoadBalancerFailed"
)

func (c *Controller) syncServiceStatus(
	service *latticev1.Service,
	deployment *appsv1.Deployment,
	kubeService *corev1.Service,
	nodePool *latticev1.NodePool,
	serviceAddress *latticev1.ServiceAddress,
	loadBalancer *latticev1.LoadBalancer,
	loadBalancerNeeded bool,
) (*latticev1.Service, error) {
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
	var state latticev1.ServiceState
	if updatedInstances < totalInstances {
		// The updated pods have not yet all been created
		state = latticev1.ServiceStateScaling
	} else if totalInstances > updatedInstances {
		// There are extra pods still
		state = latticev1.ServiceStateScaling
	} else if availableInstances < updatedInstances {
		// there's only updated instances but there aren't enough available instances yet
		state = latticev1.ServiceStateScaling
	} else if updatedInstances < desiredInstances {
		// there only exists updatedInstances, and they're all available,
		// but there isn't enough of them yet
		state = latticev1.ServiceStateScaling
	} else {
		// there are enough available updated instances, and no other instances
		state = latticev1.ServiceStateStable
	}

	// If the Deployment controller hasn't yet seen the update, it's updating
	if deployment.Generation > deployment.Status.ObservedGeneration {
		state = latticev1.ServiceStateUpdating
	}

	// If we have any stale instances though, we are updating (which can include scaling)
	// An updating status takes priority over a scaling/stable state
	if staleInstances != 0 {
		state = latticev1.ServiceStateUpdating
	}

	// The cloud controller is responsible for creating the Kubernetes Service.
	if kubeService == nil {
		state = latticev1.ServiceStateUpdating
	}

	publicPorts := latticev1.ServiceStatusPublicPorts{}
	if loadBalancerNeeded {
		switch loadBalancer.Status.State {
		case latticev1.LoadBalancerStatePending, latticev1.LoadBalancerStateProvisioning:
			state = latticev1.ServiceStateUpdating

		case latticev1.LoadBalancerStateCreated:
			for port, portInfo := range loadBalancer.Status.Ports {
				publicPorts[port] = latticev1.ServiceStatusPublicPort{
					Address: portInfo.Address,
				}
			}

		case latticev1.LoadBalancerStateFailed:
			// Only create new failure info if we didn't fail above
			if !failed {
				now := metav1.Now()
				failed = true
				failureReason = reasonLoadBalancerFailed
				failureMessage = ""
				failureTime = &now
			}

		default:
			err := fmt.Errorf(
				"LoadBalancer %v/%v has unexpected state %v",
				loadBalancer.Namespace,
				loadBalancer.Name,
				loadBalancer.Status.State,
			)
			return nil, err
		}
	}

	// But if we have a failure, our updating or scaling has failed
	// A failed status takes priority over an updating status
	var failureInfo *latticev1.ServiceFailureInfo
	if failed {
		state = latticev1.ServiceStateFailed
		switch failureReason {
		case reasonTimedOut:
			failureInfo = &latticev1.ServiceFailureInfo{
				Internal: false,
				Message:  "timed out",
				Time:     *failureTime,
			}

		case reasonLoadBalancerFailed:
			failureInfo = &latticev1.ServiceFailureInfo{
				Internal: false,
				Message:  "load balancer failed",
				Time:     *failureTime,
			}

		default:
			failureInfo = &latticev1.ServiceFailureInfo{
				Internal: true,
				Message:  fmt.Sprintf("%v: %v", failureReason, failureMessage),
				Time:     *failureTime,
			}
		}
	}

	if state == latticev1.ServiceStateStable {
		// If we still think that the deployment is stable, check to see if there are any pods
		// for this service that are terminating.
		//
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

		for _, pod := range pods {
			if pod.DeletionTimestamp != nil {
				state = latticev1.ServiceStateScaling
				staleInstances++
			}
		}
	}

	return c.updateServiceStatus(service, state, updatedInstances, staleInstances, publicPorts, failureInfo)
}

func (c *Controller) updateServiceStatus(
	service *latticev1.Service,
	state latticev1.ServiceState,
	updatedInstances, staleInstances int32,
	publicPorts latticev1.ServiceStatusPublicPorts,
	failureInfo *latticev1.ServiceFailureInfo,
) (*latticev1.Service, error) {
	status := latticev1.ServiceStatus{
		State:              state,
		ObservedGeneration: service.Generation,
		UpdateProcessed:    true,
		UpdatedInstances:   updatedInstances,
		StaleInstances:     staleInstances,
		PublicPorts:        publicPorts,
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

type lookupDelete struct {
	lookup func() (interface{}, error)
	delete func() error
}

func (c *Controller) syncDeletedService(service *latticev1.Service) error {
	lookupDeletes := []lookupDelete{
		// node pool
		// FIXME: need to change this to support system etc level node pools
		// FIXME: should potentially wait until deployment is cleaned up before deleting node pool
		//        to allow for graceful termination
		{
			lookup: func() (interface{}, error) {
				return c.nodePoolLister.NodePools(service.Namespace).Get(service.Name)
			},
			delete: func() error {
				return c.latticeClient.LatticeV1().NodePools(service.Namespace).Delete(service.Name, nil)
			},
		},
		// deployment
		// FIXME: is any of this even working? we don't name the deployment this
		{
			lookup: func() (interface{}, error) {
				return c.deploymentLister.Deployments(service.Namespace).Get(service.Name)
			},
			delete: func() error {
				return c.kubeClient.AppsV1().Deployments(service.Namespace).Delete(service.Name, nil)
			},
		},
		// kube service
		{
			lookup: func() (interface{}, error) {
				name := kubeutil.GetKubeServiceNameForService(service.Name)
				return c.kubeServiceLister.Services(service.Namespace).Get(name)
			},
			delete: func() error {
				name := kubeutil.GetKubeServiceNameForService(service.Name)
				return c.kubeClient.CoreV1().Services(service.Namespace).Delete(name, nil)
			},
		},
		// service address
		{
			lookup: func() (interface{}, error) {
				return c.serviceAddressLister.ServiceAddresses(service.Namespace).Get(service.Name)
			},
			delete: func() error {
				return c.latticeClient.LatticeV1().ServiceAddresses(service.Namespace).Delete(service.Name, nil)
			},
		},
		// load balancer
		{
			lookup: func() (interface{}, error) {
				return c.loadBalancerLister.LoadBalancers(service.Namespace).Get(service.Name)
			},
			delete: func() error {
				return c.latticeClient.LatticeV1().LoadBalancers(service.Namespace).Delete(service.Name, nil)
			},
		},
	}

	existingResource := false
	for _, lookupDelete := range lookupDeletes {
		exists, err := resourceExists(lookupDelete.lookup)
		if err != nil {
			return err
		}

		if exists {
			existingResource = true
			if err := lookupDelete.delete(); err != nil {
				return err
			}

			continue
		}
	}

	if existingResource {
		return nil
	}

	// All of the children resources have been cleaned up
	_, err := c.removeFinalizer(service)
	return err
}

func resourceExists(lookupFunc func() (interface{}, error)) (bool, error) {
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

func (c *Controller) addFinalizer(service *latticev1.Service) (*latticev1.Service, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range service.Finalizers {
		if finalizer == finalizerName {
			glog.V(5).Infof("service %v has %v finalizer", service.Name, finalizerName)
			return service, nil
		}
	}

	// Add the finalizer to the list and update.
	// If this fails due to a race the Endpoint should get requeued by the controller, so
	// not a big deal.
	service.Finalizers = append(service.Finalizers, finalizerName)
	glog.V(5).Infof("service %v missing %v finalizer, adding it", service.Name, finalizerName)

	return c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
}

func (c *Controller) removeFinalizer(service *latticev1.Service) (*latticev1.Service, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	found := false
	var finalizers []string
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