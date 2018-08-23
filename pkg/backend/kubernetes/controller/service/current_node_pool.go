package service

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Controller) syncCurrentNodePool(service *latticev1.Service) (*latticev1.NodePool, error) {
	info, err := c.nodePoolInfo(service)
	if err != nil {
		return nil, err
	}

	switch info.nodePoolType {
	case latticev1.NodePoolTypeServiceDedicated:
		return c.syncDedicatedNodePool(service, info.numInstances, info.instanceType)

	case latticev1.NodePoolTypeSystemShared:
		return c.syncSharedNodePool(service.Namespace, info.path)

	default:
		err := fmt.Errorf("unrecognized node pool type for %v: %v", service.Description(c.namespacePrefix), info.nodePoolType)
		return nil, err
	}
}

// syncDedicatedNodePool checks to see if a node pool dedicated to running a single instance of this
// service on each node exists. if it does not exist, it creates one. if it does exist, it updates it if the update
// is one that can be done in place (e.g. scaling), or creates a new one if it requires a rolling update (e.g. instance type change)
func (c *Controller) syncDedicatedNodePool(service *latticev1.Service, numInstances int32, instanceType string) (*latticev1.NodePool, error) {
	nodePool, err := c.dedicatedNodePool(service)
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

func (c *Controller) syncSharedNodePool(namespace string, path v1.NodePoolPath) (*latticev1.NodePool, error) {
	if path.Name == nil {
		return nil, fmt.Errorf("expected shared node pool path to have name, only has path %v", path.Path.String())
	}

	selector, err := sharedNodePoolSelector(namespace, path.Path, *path.Name)
	if err != nil {
		return nil, err
	}

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

func sharedNodePoolSelector(namespace string, path tree.Path, name string) (labels.Selector, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(
		latticev1.NodePoolSystemSharedPathLabelKey,
		selection.Equals,
		[]string{path.ToDomain()},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting selector for cached node pool %v:%v in namespace %v", path.String(), name, namespace)
	}
	selector = selector.Add(*requirement)

	requirement, err = labels.NewRequirement(
		latticev1.NodePoolSystemSharedNameLabelKey,
		selection.Equals,
		[]string{name},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting selector for cached node pool %v in namespace %v", path.String(), namespace)
	}
	selector = selector.Add(*requirement)

	return selector, nil
}
