package service

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/satori/go.uuid"
	"reflect"
)

// dedicatedNodePool returns a dedicated node pool for the service that has the same instanceType.
// Will return an error if it finds multiple node pools with the same instance type.
// Returns nil, nil if it could not find a matching node pool.
func (c *Controller) dedicatedNodePool(service *latticev1.Service) (*latticev1.NodePool, error) {
	// First check the cache to see if there are any matching node pools
	nodePool, err := c.cachedDedicatedNodePool(service)
	if err != nil {
		return nil, err
	}

	if nodePool != nil {
		return nodePool, nil
	}

	// Couldn't find a matching node pool in the cache. This likely means one doesn't exist,
	// but because we shouldn't orphan node pools, we need to do a quorom read from the API
	// to ensure a matching node pool does not exist before we can create a new one
	return c.quorumDedicatedNodePools(service)
}

func (c *Controller) cachedDedicatedNodePool(service *latticev1.Service) (*latticev1.NodePool, error) {
	selector, err := serviceNodePoolSelector(service)
	if err != nil {
		return nil, err
	}

	nodePools, err := c.nodePoolLister.NodePools(service.Namespace).List(selector)
	if err != nil {
		err := fmt.Errorf(
			"error trying to get cached dedicated node pool for %v: %v",
			service.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	if len(nodePools) == 0 {
		return nil, nil
	}

	if len(nodePools) == 1 {
		nodePool := nodePools[0]
		if nodePool.DeletionTimestamp != nil {
			return nil, nil
		}

		return nodePool, nil
	}

	var nodePool *latticev1.NodePool
	for _, np := range nodePools {
		if np.DeletionTimestamp != nil {
			continue
		}

		if nodePool != nil {
			err := fmt.Errorf(
				"found multiple matching cached dedicated node pools for %v: at least %v and %v",
				service.Description(c.namespacePrefix),
				nodePool.Description(c.namespacePrefix),
				np.Description(c.namespacePrefix),
			)
			return nil, err
		}

		nodePool = np
	}

	return nodePool, nil
}

func (c *Controller) quorumDedicatedNodePools(service *latticev1.Service) (*latticev1.NodePool, error) {
	selector, err := serviceNodePoolSelector(service)
	if err != nil {
		return nil, err
	}

	nodePoolList, err := c.latticeClient.LatticeV1().NodePools(service.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		err := fmt.Errorf(
			"error trying to get dedicated node pool for %v: %v",
			service.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	if len(nodePoolList.Items) == 0 {
		return nil, nil
	}

	if len(nodePoolList.Items) == 1 {
		nodePool := &nodePoolList.Items[0]
		if nodePool.DeletionTimestamp != nil {
			return nil, nil
		}

		return nodePool, nil
	}

	var nodePool *latticev1.NodePool
	for _, np := range nodePoolList.Items {
		if np.DeletionTimestamp != nil {
			continue
		}

		if nodePool != nil {
			err := fmt.Errorf(
				"found multiple matching dedicated node pools for %v: at least %v and %v",
				service.Description(c.namespacePrefix),
				nodePool.Description(c.namespacePrefix),
				np.Description(c.namespacePrefix),
			)
			return nil, err
		}

		nodePool = &np
	}

	return nodePool, nil
}

func serviceNodePoolSelector(service *latticev1.Service) (labels.Selector, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.NodePoolServiceDedicatedIDLabelKey, selection.Equals, []string{service.Name})
	if err != nil {
		return nil, err
	}

	selector = selector.Add(*requirement)
	return selector, nil
}

func (c *Controller) nodePoolServices(nodePool *latticev1.NodePool) ([]latticev1.Service, error) {
	if serviceID, ok := nodePool.ServiceDedicatedIDLabel(); ok {
		service, err := c.serviceLister.Services(nodePool.Namespace).Get(serviceID)
		if err != nil {
			return nil, err
		}

		return []latticev1.Service{*service}, nil
	}

	_, ok, err := nodePool.SystemSharedPathLabel()
	if err != nil {
		return nil, err
	}
	if ok {
		services, err := c.serviceLister.Services(nodePool.Namespace).List(labels.Everything())
		if err != nil {
			return nil, err
		}

		var nodePoolServices []latticev1.Service
		for _, service := range services {
			nodePools, err := service.NodePoolAnnotation()
			if err != nil {
				// FIXME: log/send warn event
				continue
			}

			if nodePools.ContainsNodePool(nodePool.Namespace, nodePool.Name) {
				nodePoolServices = append(nodePoolServices, *service)
			}
		}
		return nodePoolServices, nil
	}

	err = fmt.Errorf(
		"%v did not have %v or %v annotation",
		nodePool.Description(c.namespacePrefix),
		latticev1.NodePoolServiceDedicatedIDLabelKey,
		latticev1.NodePoolSystemSharedPathLabelKey,
	)
	return nil, err
}

func (c *Controller) updateNodePoolSpec(nodePool *latticev1.NodePool, desiredSpec latticev1.NodePoolSpec) (*latticev1.NodePool, error) {
	if reflect.DeepEqual(nodePool.Spec, desiredSpec) {
		return nodePool, nil
	}

	// Copy so the shared cache isn't mutated
	nodePool = nodePool.DeepCopy()
	nodePool.Spec = desiredSpec

	result, err := c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
	if err != nil {
		err := fmt.Errorf("error trying to update %v spec: %v", nodePool.Description(c.namespacePrefix), err)
		return nil, err
	}

	return result, nil
}

func (c *Controller) createNewDedicatedNodePool(service *latticev1.Service, numInstances int32, instanceType string) (*latticev1.NodePool, error) {
	nodePool := newDedicatedNodePool(service, numInstances, instanceType)
	result, err := c.latticeClient.LatticeV1().NodePools(service.Namespace).Create(nodePool)
	if err != nil {
		err := fmt.Errorf(
			"error trying to create new dedicated node pool for %v: %v",
			service.Description(c.namespacePrefix),
			err,
		)
		return nil, err
	}

	return result, nil
}

func newDedicatedNodePool(service *latticev1.Service, numInstances int32, instanceType string) *latticev1.NodePool {
	spec := nodePoolSpec(numInstances, instanceType)

	nodePool := &latticev1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:            uuid.NewV4().String(),
			OwnerReferences: []metav1.OwnerReference{*controllerRef(service)},
			Labels: map[string]string{
				latticev1.NodePoolServiceDedicatedIDLabelKey: service.Name,
			},
		},
		Spec: spec,
	}
	return nodePool
}

func nodePoolSpec(numInstances int32, instanceType string) latticev1.NodePoolSpec {
	return latticev1.NodePoolSpec{
		NumInstances: numInstances,
		InstanceType: instanceType,
	}
}

func (c *Controller) currentEpochStable(nodePool *latticev1.NodePool) (bool, error) {
	if !nodePool.UpdateProcessed() {
		return false, nil
	}

	currentEpoch, ok := nodePool.Status.Epochs.CurrentEpoch()
	if !ok {
		err := fmt.Errorf(
			"%v for %v is processed but does not have a current epoch",
			nodePool.Description(c.namespacePrefix),
		)
		return false, err
	}

	epochStatus, ok := nodePool.Status.Epochs.Epoch(currentEpoch)
	if !ok {
		err := fmt.Errorf(
			"%v claims to have current epoch %v but does not have a status for it",
			nodePool.Description(c.namespacePrefix),
			currentEpoch,
		)
		return false, err
	}

	return epochStatus.State == latticev1.NodePoolStateStable, nil
}
