package service

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Controller) cleanUpDedicatedNodePool(
	service *latticev1.Service,
	currentNodePool *latticev1.NodePool,
	deploymentStatus *deploymentStatus,
) (bool, error) {
	// If the deployment is ready, that means that it is fully stable on the most up to date
	// epoch of the correct node pool, and no pods are running on any other node pool epochs.
	if !deploymentStatus.Ready() {
		return false, nil
	}

	// TODO: should formally think about if it's possible we'll miss a node pool here.
	// intuitively, i don't think we should, since if we've seen our current node pool be ready,
	// the cache should contain any node pools that are older than it
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.NodePoolServiceDedicatedIDLabelKey, selection.Equals, []string{service.Name})
	if err != nil {
		return false, err
	}
	selector = selector.Add(*requirement)

	nodePools, err := c.nodePoolLister.NodePools(service.Namespace).List(selector)
	if err != nil {
		err := fmt.Errorf(
			"error trying to get cached dedicated node pool for %v: %v",
			service.Description(c.namespacePrefix),
			err,
		)
		return false, err
	}

	extraNodePoolsExist := false
	for _, nodePool := range nodePools {
		if nodePool.UID == currentNodePool.UID {
			continue
		}

		extraNodePoolsExist = true

		if currentNodePool.DeletionTimestamp != nil {
			continue
		}

		err = c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Delete(nodePool.Name, nil)
		if err != nil {
			err := fmt.Errorf(
				"error trying to delete extra %v for %v: %v",
				nodePool.Description(c.namespacePrefix),
				service.Description(c.namespacePrefix),
				err,
			)
			return false, err
		}
	}

	return extraNodePoolsExist, nil
}
