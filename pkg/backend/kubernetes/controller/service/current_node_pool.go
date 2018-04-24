package service

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/satori/go.uuid"
)

func (c *Controller) syncCurrentNodePool(service *latticev1.Service) (*latticev1.NodePool, error) {
	resources := service.Spec.Definition.Resources()

	var numInstances int32
	if service.Spec.Definition.Resources().NumInstances != nil {
		numInstances = *service.Spec.Definition.Resources().NumInstances
	} else if service.Spec.Definition.Resources().MinInstances != nil {
		numInstances = *service.Spec.Definition.Resources().MinInstances
	} else {
		return nil, fmt.Errorf("%v did not specify num instances or min instances", service.Description(c.namespacePrefix))
	}

	var instanceType string
	if resources.NodePool != nil {
		if resources.NodePool.NodePoolName != nil {
			nodePoolPath, err := tree.NewNodePath(*resources.NodePool.NodePoolName)
			if err != nil {
				return nil, fmt.Errorf("error parsing shared node pool path: %v", err)
			}

			return c.syncSharedNodePool(service.Namespace, nodePoolPath)
		}

		if resources.NodePool.NodePool == nil {
			return nil, fmt.Errorf("%v has non-null node pool block, but does not specify node pool name or node pool", service)
		}

		instanceType = resources.NodePool.NodePool.InstanceType
	} else {
		if resources.InstanceType == nil {
			return nil, fmt.Errorf("%v did not specify a node pool or instance type", service.Description(c.namespacePrefix))
		}
		instanceType = *resources.InstanceType
	}

	return c.syncDedicatedNodePool(service, numInstances, instanceType)
}

// syncDedicatedNodePool checks to see if a node pool dedicated to running a single instance of this
// service on each node exists. if it does not exist, it creates one. if it does exist, it updates it if the update
// is one that can be done in place (e.g. scaling), or creates a new one if it requires a rolling update (e.g. instance type change)
func (c *Controller) syncDedicatedNodePool(service *latticev1.Service, numInstances int32, instanceType string) (*latticev1.NodePool, error) {
	nodePool, err := c.dedicatedNodePool(service, instanceType)
	if err != nil {
		return nil, err
	}

	// We didn't find a matching dedicated node, so we'll make a new one
	if nodePool == nil {
		nodePool, err := c.createNewDedicatedNodePool(service, numInstances, instanceType)
		if err != nil {
			return nil, err
		}

		return nodePool, nil
	}

	nodePool, err = c.syncExistingDedicatedNodePool(nodePool, numInstances, instanceType)
	return nodePool, err
}

func (c *Controller) syncSharedNodePool(namespace string, path tree.NodePath) (*latticev1.NodePool, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.NodePoolSystemSharedPathLabelKey, selection.Equals, []string{path.ToDomain()})
	if err != nil {
		return nil, err
	}

	selector = selector.Add(*requirement)
	nodePools, err := c.nodePoolLister.NodePools(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	// the shared node pool doesn't exist yet
	if len(nodePools) == 0 {
		return nil, nil
	}

	if len(nodePools) > 1 {
		// FIXME: send warning or something
		return nil, fmt.Errorf("found multiple node pools matching path %v in namespace %v", path.String(), namespace)
	}

	nodePool := nodePools[0]
	return nodePool, nil
}

func (c *Controller) syncExistingDedicatedNodePool(nodePool *latticev1.NodePool, numInstances int32, instanceType string) (*latticev1.NodePool, error) {
	spec := nodePoolSpec(numInstances, instanceType)
	return c.updateNodePoolSpec(nodePool, spec)
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
		Status: latticev1.NodePoolStatus{
			State: latticev1.NodePoolStatePending,
		},
	}
	return nodePool
}

func nodePoolSpec(numInstances int32, instanceType string) latticev1.NodePoolSpec {
	return latticev1.NodePoolSpec{
		NumInstances: numInstances,
		InstanceType: instanceType,
	}
}
