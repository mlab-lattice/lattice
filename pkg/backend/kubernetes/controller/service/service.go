package service

import (
	"encoding/json"
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	reasonTimedOut           = "ProgressDeadlineExceeded"
	reasonLoadBalancerFailed = "LoadBalancerFailed"
)

func (c *Controller) syncServiceStatus(
	service *latticev1.Service,
	nodePool *latticev1.NodePool,
	address *latticev1.Address,
	deploymentStatus *deploymentStatus,
	extraNodePoolsExist bool,
) (*latticev1.Service, error) {
	failed := false
	failureReason := ""
	var failureTime *metav1.Time

	var state latticev1.ServiceState
	if !deploymentStatus.UpdateProcessed {
		state = latticev1.ServiceStateUpdating
	} else if deploymentStatus.State == deploymentStatePending {
		state = latticev1.ServiceStateUpdating
	} else if deploymentStatus.State == deploymentStateFailed {
		failed = true
		if deploymentStatus.FailureInfo != nil {
			failureReason = deploymentStatus.FailureInfo.Reason
			failureTime = &deploymentStatus.FailureInfo.Time
		}
	} else if deploymentStatus.State == deploymentStateScaling {
		state = latticev1.ServiceStateScaling
	} else {
		state = latticev1.ServiceStateStable
	}

	// But if we have a failure, our updating or scaling has failed
	// A failed status takes priority over an updating status
	var failureInfo *latticev1.ServiceStatusFailureInfo
	if failed {
		state = latticev1.ServiceStateFailed
		if failureTime == nil {
			now := metav1.Now()
			failureTime = &now
		}

		switch failureReason {
		case reasonTimedOut:
			failureInfo = &latticev1.ServiceStatusFailureInfo{
				Internal: false,
				Message:  "timed out",
				Time:     *failureTime,
			}

		case reasonLoadBalancerFailed:
			failureInfo = &latticev1.ServiceStatusFailureInfo{
				Internal: false,
				Message:  "load balancer failed",
				Time:     *failureTime,
			}

		default:
			failureInfo = &latticev1.ServiceStatusFailureInfo{
				Internal: true,
				Message:  failureReason,
				Time:     *failureTime,
			}
		}
	}

	// we only update the deployment spec once the node pool is stable,
	// so if it is not stable we don't need to update the service's node
	// pool annotation
	if nodePool.Stable() {
		var err error
		service, err = c.updateServiceNodePoolAnnotation(service, nodePool, state)
		if err != nil {
			return nil, err
		}
	}

	return c.updateServiceStatus(
		service,
		state,
		deploymentStatus.Reason,
		failureInfo,
		deploymentStatus.AvailableInstances,
		deploymentStatus.UpdatedInstances,
		deploymentStatus.StaleInstances,
		deploymentStatus.TerminatingInstances,
		address.Status.Ports,
	)
}

func (c *Controller) updateServiceNodePoolAnnotation(
	service *latticev1.Service,
	nodePool *latticev1.NodePool,
	state latticev1.ServiceState,
) (*latticev1.Service, error) {
	newAnnotation := make(latticev1.NodePoolAnnotationValue)
	existingAnnotation, err := service.NodePoolAnnotation()
	if err != nil {
		err := fmt.Errorf("error getting existing node pool annotation for %v: %v", service.Description(c.namespacePrefix), err)
		return nil, err
	}

	// If the service is currently stable, then we are only running on the
	// current epoch of the current node pool. If it's not stable we can't
	// assume that we're fully off of previous node pools and epochs, so
	// we have to include the values from the existing annotation.
	if state != latticev1.ServiceStateStable {
		newAnnotation = existingAnnotation
	}

	epoch, ok := nodePool.Status.Epochs.CurrentEpoch()
	if !ok {
		return nil, fmt.Errorf("%v is stable but does not have a current epoch", nodePool.Description(c.namespacePrefix))
	}

	newAnnotation.Add(nodePool.Namespace, nodePool.Name, epoch)

	if reflect.DeepEqual(existingAnnotation, newAnnotation) {
		return service, nil
	}

	newAnnotationJSON, err := json.Marshal(&newAnnotation)
	if err != nil {
		return nil, fmt.Errorf("error marshalling node pool annotation: %v", err)
	}

	// Copy the service so the shared cache isn't mutated
	service = service.DeepCopy()
	service.Annotations[latticev1.NodePoolWorkloadAnnotationKey] = string(newAnnotationJSON)

	result, err := c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
	if err != nil {
		return nil, fmt.Errorf("error updating %v node pool annotation: %v", service.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) updateServiceStatus(
	service *latticev1.Service,
	state latticev1.ServiceState,
	reason *string,
	failureInfo *latticev1.ServiceStatusFailureInfo,
	availableInstances, updatedInstances, staleInstances, terminatingInstances int32,
	ports map[int32]string,
) (*latticev1.Service, error) {
	status := latticev1.ServiceStatus{
		ObservedGeneration: service.Generation,

		State:       state,
		Reason:      reason,
		FailureInfo: failureInfo,

		AvailableInstances:   availableInstances,
		UpdatedInstances:     updatedInstances,
		StaleInstances:       staleInstances,
		TerminatingInstances: terminatingInstances,

		Ports: ports,
	}

	if reflect.DeepEqual(service.Status, status) {
		return service, nil
	}

	// Copy the service so the shared cache isn't mutated
	service = service.DeepCopy()
	service.Status = status

	result, err := c.latticeClient.LatticeV1().Services(service.Namespace).UpdateStatus(service)
	if err != nil {
		return nil, fmt.Errorf("error updating status for %v: %v", service.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func controllerRef(service *latticev1.Service) *metav1.OwnerReference {
	return metav1.NewControllerRef(service, latticev1.ServiceKind)
}
