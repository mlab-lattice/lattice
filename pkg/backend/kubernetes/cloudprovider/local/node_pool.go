package local

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (cp *DefaultLocalCloudProvider) NodePoolNeedsNewEpoch(nodePool *latticev1.NodePool) (bool, error) {
	// Just need a single epoch to appease the node controller
	_, ok := nodePool.Status.Epochs.CurrentEpoch()
	return !ok, nil
}

func (cp *DefaultLocalCloudProvider) NodePoolCurrentEpochState(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
) (latticev1.NodePoolState, error) {
	current, ok := nodePool.Status.Epochs.CurrentEpoch()
	if !ok {
		return latticev1.NodePoolStatePending, fmt.Errorf("could not get current epoch for %v", nodePool.Description(cp.namespacePrefix))
	}

	epochInfo, ok := nodePool.Status.Epochs.Epoch(current)
	if !ok {
		return latticev1.NodePoolStatePending, fmt.Errorf("could not get epoch status for %v epoch %v", nodePool.Description(cp.namespacePrefix), current)
	}

	return epochInfo.State, nil
}

func (cp *DefaultLocalCloudProvider) NodePoolAddAnnotations(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	annotations map[string]string,
	epoch latticev1.NodePoolEpoch,
) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) ProvisionNodePoolEpoch(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) DeprovisionNodePoolEpoch(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
) error {
	return nil
}
