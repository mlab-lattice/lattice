package service

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/satori/go.uuid"
)

// dedicatedNodePool returns a dedicated node pool for the service that has the same instanceType.
// Will return an error if it finds multiple node pools with the same instance type.
// Returns nil, nil if it could not find a matching node pool.
func (c *Controller) dedicatedNodePool(service *latticev1.Service, instanceType string) (*latticev1.NodePool, error) {
	// First check the cache to see if there are any matching node pools
	cachedNodePools, err := c.cachedDedicatedNodePools(service)
	if err != nil {
		return nil, err
	}

	// Try to find a node pool that can be synced in place.
	// Currently, this means the instance types match
	var matchingNodePool *latticev1.NodePool
	for _, nodePool := range cachedNodePools {
		if nodePool.Spec.InstanceType == instanceType {
			if nodePool.DeletionTimestamp != nil {
				return nil, fmt.Errorf("found %v for %v but it is being deleted", nodePool.Description(c.namespacePrefix), service.Description(c.namespacePrefix))
			}

			if matchingNodePool != nil {
				err := fmt.Errorf(
					"found multiple identical dedicated node pools (at least %v and %v) for service %v",
					matchingNodePool.Description(c.namespacePrefix),
					nodePool.Description(c.namespacePrefix),
					service.Description(c.namespacePrefix),
				)
				return nil, err
			}

			matchingNodePool = nodePool
		}
	}

	if matchingNodePool != nil {
		return matchingNodePool, nil
	}

	// Couldn't find a matching node pool in the cache. This likely means one doesn't exist,
	// but because we shouldn't orphan node pools, we need to do a quorom read from the API
	// to ensure a matching node pool does not exist before we can create a new one
	nodePools, err := c.quorumDedicatedNodePools(service)
	if err != nil {
		return nil, err
	}

	for _, nodePool := range nodePools {
		if nodePool.Spec.InstanceType == instanceType {
			if nodePool.DeletionTimestamp != nil {
				return nil, fmt.Errorf("found %v for %v but it is being deleted", nodePool.Description(c.namespacePrefix), service.Description(c.namespacePrefix))
			}

			if matchingNodePool != nil {
				err := fmt.Errorf(
					"found multiple identical dedicated node pools (at least %v and %v) for service %v",
					matchingNodePool.Description(c.namespacePrefix),
					nodePool.Description(c.namespacePrefix),
					service.Description(c.namespacePrefix),
				)
				return nil, err
			}

			matchingNodePool = &nodePool
		}
	}

	return matchingNodePool, nil
}

func (c *Controller) cachedDedicatedNodePools(service *latticev1.Service) ([]*latticev1.NodePool, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.NodePoolServiceDedicatedIDLabelKey, selection.Equals, []string{service.Name})
	if err != nil {
		return nil, err
	}

	selector = selector.Add(*requirement)
	return c.nodePoolLister.NodePools(service.Namespace).List(selector)
}

func (c *Controller) quorumDedicatedNodePools(service *latticev1.Service) ([]latticev1.NodePool, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.NodePoolServiceDedicatedIDLabelKey, selection.Equals, []string{service.Name})
	if err != nil {
		return nil, err
	}

	selector = selector.Add(*requirement)
	list, err := c.latticeClient.LatticeV1().NodePools(service.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
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
	// Copy so the shared cache isn't mutated
	nodePool = nodePool.DeepCopy()
	nodePool.Spec = desiredSpec

	return c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Update(nodePool)
}

func (c *Controller) createNewDedicatedNodePool(service *latticev1.Service, numInstances int32, instanceType string) (*latticev1.NodePool, error) {
	nodePool := newDedicatedNodePool(service, numInstances, instanceType)
	return c.latticeClient.LatticeV1().NodePools(service.Namespace).Create(nodePool)
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
