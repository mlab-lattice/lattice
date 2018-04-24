package service

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) cleanUpDedicatedNodePool(
	service *latticev1.Service,
	nodePool *latticev1.NodePool,
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
	cachedNodePools, err := c.cachedDedicatedNodePools(service)
	if err != nil {
		return false, err
	}

	extraNodePoolsExist := false
	for _, cachedNodePool := range cachedNodePools {
		if cachedNodePool.UID == nodePool.UID {
			continue
		}

		if nodePool.DeletionTimestamp != nil {
			extraNodePoolsExist = true
		}

		// TODO: send event
		err = c.latticeClient.LatticeV1().NodePools(cachedNodePool.Namespace).Delete(cachedNodePool.Name, nil)
		if err != nil {
			return false, err
		}

		extraNodePoolsExist = true
	}

	return extraNodePoolsExist, nil
}
