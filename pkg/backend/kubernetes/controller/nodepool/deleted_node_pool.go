package nodepool

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncDeletedNodePool(nodePool *latticev1.NodePool) error {
	nodePool, err := c.retireEpochs(nodePool, true)
	if err != nil {
		return err
	}

	// If there are still epochs that weren't able to be retired yet, we are done for this round.
	// Workloads will be terminated and this node pool will be requeued and reassessed.
	if len(nodePool.Status.Epochs) > 0 {
		return nil
	}

	// If all of the epochs were retired, the node pool has been deleted so we can remove the finalizer.
	_, err = c.removeFinalizer(nodePool)
	return err
}
