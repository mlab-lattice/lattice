package job

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Controller) nodePool(job *latticev1.Job) (*latticev1.NodePool, error) {
	nodePoolPath := job.Spec.Definition.NodePool
	selector, err := sharedNodePoolSelector(job.Namespace, nodePoolPath.Path(), nodePoolPath.Subcomponent())
	if err != nil {
		return nil, err
	}

	nodePools, err := c.nodePoolLister.NodePools(job.Namespace).List(selector)
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
			nodePoolPath.String(),
			job.Namespace,
		)
		return nil, err
	}

	return nodePools[0], nil
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
