package local

import (
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (cp *DefaultLocalCloudProvider) ProvisionNodePool(latticeID v1.LatticeID, nodePool *latticev1.NodePool) (*latticev1.NodePool, *time.Duration, error) {
	return nodePool, nil, nil
}

func (cp *DefaultLocalCloudProvider) DeprovisionNodePool(latticeID v1.LatticeID, nodePool *latticev1.NodePool) (*time.Duration, error) {
	return nil, nil
}

func (cp *DefaultLocalCloudProvider) NodePoolState(latticeID v1.LatticeID, nodePool *latticev1.NodePool) (latticev1.NodePoolState, error) {
	return nodePool.Status.State, nil
}
