package service

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) deleteExtraNodePools(
	service *latticev1.Service,
	nodePool *latticev1.NodePool,
	nodePoolReady bool,
	deploymentStatus *deploymentStatus,
) (bool, error) {
	// Need to wait until the current node pool for the deployment is ready, and the deployment is
	// stable on that node pool before we can clean up old node pools.
	if !nodePoolReady || !deploymentStatus.UpdateProcessed || deploymentStatus.State == deploymentStateStable {
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
