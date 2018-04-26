package service

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
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

	return c.syncExistingDedicatedNodePool(nodePool, numInstances, instanceType)
}

func (c *Controller) syncSharedNodePool(namespace string, path tree.NodePath) (*latticev1.NodePool, error) {
	// TODO: how to handle a shared node pool move?
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
		err := fmt.Errorf(
			"found multiple shared node pools matching path %v in namespace %v",
			path.String(),
			namespace,
		)
		return nil, err
	}

	return nodePools[0], nil
}

func (c *Controller) syncExistingDedicatedNodePool(nodePool *latticev1.NodePool, numInstances int32, instanceType string) (*latticev1.NodePool, error) {
	spec := nodePoolSpec(numInstances, instanceType)
	return c.updateNodePoolSpec(nodePool, spec)
}
