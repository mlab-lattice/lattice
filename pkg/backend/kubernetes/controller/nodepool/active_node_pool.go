package nodepool

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncActiveNodePool(nodePool *latticev1.NodePool) error {
	// Get the status of existing epochs.
	epochs := make(latticev1.NodePoolStatusEpochs)
	for _, epoch := range nodePool.Status.Epochs.Epochs() {
		epochInfo, ok := nodePool.Status.Epochs.Epoch(epoch)
		if !ok {
			return fmt.Errorf("could not get info for %v epoch %v", nodePool.Description(c.namespacePrefix), epoch)
		}

		status, err := c.cloudProvider.NodePoolEpochStatus(c.latticeID, nodePool, epoch, &epochInfo.Spec)
		if err != nil {
			return fmt.Errorf(
				"error getting status for %v epoch %v: %v",
				nodePool.Description(c.namespacePrefix),
				epoch,
				err,
			)
		}

		epochs[epoch] = latticev1.NodePoolStatusEpoch{
			Spec:   epochInfo.Spec,
			Status: *status,
		}
	}

	needsNewEpoch, err := c.cloudProvider.NodePoolNeedsNewEpoch(nodePool)
	if err != nil {
		return err
	}

	// If the node pool needs a new epoch, generate it and add it to the status.
	// Otherwise just get the current epoch.
	var epoch latticev1.NodePoolEpoch

	if needsNewEpoch {
		epoch = nodePool.Status.Epochs.NextEpoch()
		epochs[epoch] = latticev1.NodePoolStatusEpoch{
			Spec: latticev1.NodePoolSpec{
				InstanceType: nodePool.Spec.InstanceType,
				NumInstances: nodePool.Spec.NumInstances,
			},
			Status: latticev1.NodePoolStatusEpochStatus{
				State: latticev1.NodePoolStatePending,
			},
		}
	} else {
		var ok bool
		epoch, ok = nodePool.Status.Epochs.CurrentEpoch()
		if !ok {
			return fmt.Errorf(
				"cloud provider reported that %v did not need new epoch, but it does not have a current epoch",
				nodePool.Description(c.namespacePrefix),
			)
		}
	}

	// Update the node pool's status prior to telling the cloud provider to provision the current epoch.
	nodePool, err = c.updateNodePoolStatus(nodePool, epochs)
	if err != nil {
		return err
	}

	err = c.cloudProvider.EnsureNodePoolEpoch(c.latticeID, nodePool, epoch)
	if err != nil {
		return fmt.Errorf(
			"cloud provider could not ensure %v epoch %v: %v",
			nodePool.Description(c.namespacePrefix),
			epoch,
			err,
		)
	}

	// Add any annotations needed by the cloud provider.
	// Copy annotations so cloud provider doesn't mutate the cache
	annotations := make(map[string]string)
	for k, v := range nodePool.Annotations {
		annotations[k] = v
	}

	err = c.cloudProvider.NodePoolAddAnnotations(c.latticeID, nodePool, annotations, epoch)
	if err != nil {
		return fmt.Errorf(
			"cloud provider could not get annotations for %v epoch %v: %v",
			nodePool.Description(c.namespacePrefix),
			epoch,
			err,
		)
	}

	nodePool, err = c.updateNodePoolAnnotations(nodePool, annotations)
	if err != nil {
		return fmt.Errorf("could not update %v annotations: %v", nodePool.Description(c.namespacePrefix), err)
	}

	status, err := c.cloudProvider.NodePoolEpochStatus(c.latticeID, nodePool, epoch, &nodePool.Spec)
	if err != nil {
		return fmt.Errorf(
			"error getting status for %v epoch %v: %v",
			nodePool.Description(c.namespacePrefix),
			epoch,
			err,
		)
	}

	// If we got to here, the node pool's current epoch is stable, so update the status to reflect that.
	epochs[epoch] = latticev1.NodePoolStatusEpoch{
		Spec:   nodePool.Spec,
		Status: *status,
	}
	nodePool, err = c.updateNodePoolStatus(nodePool, epochs)
	if err != nil {
		return err
	}

	// We successfully provisioned the current epoch, so we can start trying to retire
	// any earlier epochs.
	_, err = c.retireEpochs(nodePool, false)
	return err
}
