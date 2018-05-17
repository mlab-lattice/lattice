package service

import (
	"encoding/json"
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	reasonTimedOut           = "ProgressDeadlineExceeded"
	reasonLoadBalancerFailed = "LoadBalancerFailed"
)

type nodePoolInfo struct {
	nodePoolType latticev1.NodePoolType

	// dedicated node pool options
	instanceType string
	numInstances int32
	perInstance  bool

	// shared system node pool options
	path tree.NodePath
}

func (c *Controller) numInstances(service *latticev1.Service) (int32, error) {
	if service.Spec.Definition.Resources().NumInstances != nil {
		return *service.Spec.Definition.Resources().NumInstances, nil
	}

	if service.Spec.Definition.Resources().MinInstances != nil {
		return *service.Spec.Definition.Resources().MinInstances, nil
	}

	err := fmt.Errorf(
		"error getting num instances for %v: did not specify num instances or min instances",
		service.Description(c.namespacePrefix),
	)
	return 0, err
}

func (c *Controller) nodePoolInfo(service *latticev1.Service) (nodePoolInfo, error) {
	resources := service.Spec.Definition.Resources()

	// dedicated per-instance node pool
	if resources.NodePool == nil {
		if resources.InstanceType == nil {
			return nodePoolInfo{}, fmt.Errorf("%v did not specify a node pool or instance type", service.Description(c.namespacePrefix))
		}

		numInstances, err := c.numInstances(service)
		if err != nil {
			return nodePoolInfo{}, err
		}

		info := nodePoolInfo{
			nodePoolType: latticev1.NodePoolTypeServiceDedicated,
			instanceType: *resources.InstanceType,
			numInstances: numInstances,
			perInstance:  true,
		}
		return info, nil
	}

	// dedicated not per-instance node pool
	if resources.NodePool.NodePool != nil {
		info := nodePoolInfo{
			nodePoolType: latticev1.NodePoolTypeServiceDedicated,
			instanceType: resources.NodePool.NodePool.InstanceType,
			numInstances: resources.NodePool.NodePool.NumInstances,
			perInstance:  false,
		}
		return info, nil
	}

	if resources.NodePool.NodePoolName != nil {
		path, err := tree.NewNodePath(*resources.NodePool.NodePoolName)
		if err != nil {
			err := fmt.Errorf("error parsing shared node pool path for %v: %v", service.Description(c.namespacePrefix), err)
			return nodePoolInfo{}, err
		}

		info := nodePoolInfo{
			nodePoolType: latticev1.NodePoolTypeSystemShared,
			path:         path,
		}
		return info, nil
	}

	return nodePoolInfo{}, fmt.Errorf("%v did not specify a node pool or instance type", service.Description(c.namespacePrefix))
}

func (c *Controller) syncServiceStatus(
	service *latticev1.Service,
	nodePool *latticev1.NodePool,
	address *latticev1.Address,
	deploymentStatus *deploymentStatus,
	extraNodePoolsExist bool,
) (*latticev1.Service, error) {
	currentEpochStable, err := c.currentEpochStable(nodePool)
	if err != nil {
		err := fmt.Errorf(
			"error checking if current epoch for %v node pool is stable: %v",
			service.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	// we only update the deployment spec once the node pool is stable,
	// so if it is not stable we don't need to update the service's node
	// pool annotation
	if currentEpochStable {
		annotation, err := c.serviceNodePoolAnnotation(service, nodePool, deploymentStatus)
		if err != nil {
			return nil, err
		}

		service, err = c.updateServiceNodePoolAnnotation(service, annotation)
		if err != nil {
			return nil, err
		}
	}

	state, message, failureInfo := serviceStatus(nodePool, address, deploymentStatus, extraNodePoolsExist)
	return c.updateServiceStatus(
		service,
		state,
		message,
		failureInfo,
		deploymentStatus.AvailableInstances,
		deploymentStatus.UpdatedInstances,
		deploymentStatus.StaleInstances,
		deploymentStatus.TerminatingInstances,
		address.Status.Ports,
	)
}

func serviceStatus(
	nodePool *latticev1.NodePool,
	address *latticev1.Address,
	deploymentStatus *deploymentStatus,
	extraNodePoolsExist bool,
) (latticev1.ServiceState, *string, *latticev1.ServiceStatusFailureInfo) {
	if !deploymentStatus.UpdateProcessed {
		message := "waiting for update to be processed"
		return latticev1.ServiceStateUpdating, &message, nil
	}

	if deploymentStatus.State == deploymentStateFailed {
		failureInfo := &latticev1.ServiceStatusFailureInfo{
			Message:   "unknown error",
			Internal:  true,
			Timestamp: metav1.Now(),
		}

		if deploymentStatus.FailureInfo != nil {
			failureInfo.Timestamp = deploymentStatus.FailureInfo.Time

			switch deploymentStatus.FailureInfo.Reason {
			case reasonTimedOut:
				failureInfo = &latticev1.ServiceStatusFailureInfo{
					Internal: false,
					Message:  "timed out",
				}

			case reasonLoadBalancerFailed:
				failureInfo = &latticev1.ServiceStatusFailureInfo{
					Internal: false,
					Message:  "load balancer failed",
				}

			default:
				failureInfo = &latticev1.ServiceStatusFailureInfo{
					Internal: true,
					Message:  deploymentStatus.FailureInfo.Reason,
				}
			}
		}

		return latticev1.ServiceStateFailed, nil, failureInfo
	}

	if nodePool == nil {
		message := "node pool is pending"
		return latticev1.ServiceStateUpdating, &message, nil
	}

	if !nodePool.Stable() {
		message := fmt.Sprintf("node pool is %v", nodePool.Reason())
		return latticev1.ServiceStateUpdating, &message, nil
	}

	if !address.Stable() {
		message := fmt.Sprintf("address is %v", address.Reason())
		return latticev1.ServiceStateUpdating, &message, nil
	}

	if extraNodePoolsExist {
		message := "destroying stale node pools"
		return latticev1.ServiceStateUpdating, &message, nil
	}

	// this probably shouldn't happen (deployment shouldn't be pending if both
	// node pool and address are stable)
	if deploymentStatus.State == deploymentStatePending {
		message := "waiting for update to be processed"
		return latticev1.ServiceStateUpdating, &message, nil
	}

	if deploymentStatus.State == deploymentStateScaling {
		message := "instances are scaling"
		return latticev1.ServiceStateScaling, &message, nil

	}

	return latticev1.ServiceStateStable, nil, nil
}

func (c *Controller) serviceNodePoolAnnotation(
	service *latticev1.Service,
	nodePool *latticev1.NodePool,
	status *deploymentStatus,
) (latticev1.NodePoolAnnotationValue, error) {
	newAnnotation := make(latticev1.NodePoolAnnotationValue)

	// If the deployment is currently stable, then we are only running on the
	// current epoch of the current node pool. If it's not stable we can't
	// assume that we're fully off of previous node pools and epochs, so
	// we have to include the values from the existing annotation.
	if !status.Stable() {
		fmt.Println("using old annotation")
		existingAnnotation, err := service.NodePoolAnnotation()
		if err != nil {
			err := fmt.Errorf(
				"error getting existing node pool annotation for %v: %v",
				service.Description(c.namespacePrefix),
				err,
			)
			return nil, err
		}

		newAnnotation = existingAnnotation
	} else {
		fmt.Println("using new annotation")
	}

	epoch, ok := nodePool.Status.Epochs.CurrentEpoch()
	if !ok {
		return nil, fmt.Errorf("%v is stable but does not have a current epoch", nodePool.Description(c.namespacePrefix))
	}

	newAnnotation.Add(nodePool.Namespace, nodePool.Name, epoch)
	return newAnnotation, nil
}

func (c *Controller) updateServiceNodePoolAnnotation(
	service *latticev1.Service,
	annotation latticev1.NodePoolAnnotationValue,
) (*latticev1.Service, error) {
	existingAnnotation, err := service.NodePoolAnnotation()
	if err != nil {
		err := fmt.Errorf("error getting existing node pool annotation for %v: %v", service.Description(c.namespacePrefix), err)
		return nil, err
	}

	if reflect.DeepEqual(existingAnnotation, annotation) {
		return service, nil
	}

	newAnnotationJSON, err := json.Marshal(&annotation)
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
	message *string,
	failureInfo *latticev1.ServiceStatusFailureInfo,
	availableInstances, updatedInstances, staleInstances, terminatingInstances int32,
	ports map[int32]string,
) (*latticev1.Service, error) {
	status := latticev1.ServiceStatus{
		ObservedGeneration: service.Generation,

		State:       state,
		Message:     message,
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

func (c *Controller) addFinalizer(service *latticev1.Service) (*latticev1.Service, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range service.Finalizers {
		if finalizer == kubeutil.ServiceControllerFinalizer {
			return service, nil
		}
	}

	// Copy so we don't mutate the shared cache
	service = service.DeepCopy()
	service.Finalizers = append(service.Finalizers, kubeutil.ServiceControllerFinalizer)

	result, err := c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
	if err != nil {
		return nil, fmt.Errorf("error adding %v finalizer: %v", service.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) removeFinalizer(service *latticev1.Service) (*latticev1.Service, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range service.Finalizers {
		if finalizer == kubeutil.ServiceControllerFinalizer {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return service, nil
	}

	// Copy so we don't mutate the shared cache
	service = service.DeepCopy()
	service.Finalizers = finalizers

	result, err := c.latticeClient.LatticeV1().Services(service.Namespace).Update(service)
	if err != nil {
		return nil, fmt.Errorf("error removing %v finalizer: %v", service.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func controllerRef(service *latticev1.Service) *metav1.OwnerReference {
	return metav1.NewControllerRef(service, latticev1.ServiceKind)
}
