package local

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
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

func (cp *DefaultLocalCloudProvider) EnsureNodePoolEpoch(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) DestroyNodePoolEpoch(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) NodePoolEpochStatus(
	latticeID v1.LatticeID,
	nodePool *latticev1.NodePool,
	epoch latticev1.NodePoolEpoch,
	epochSpec *latticev1.NodePoolSpec,
) (*latticev1.NodePoolStatusEpoch, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.NodePoolIDLabelKey, selection.Equals, []string{nodePool.ID(epoch)})
	if err != nil {
		return nil, fmt.Errorf("error making requirement for %v node lookup: %v", nodePool.Description(cp.namespacePrefix), err)
	}

	selector = selector.Add(*requirement)
	nodes, err := cp.kubeNodeLister.List(selector)
	if err != nil {
		return nil, fmt.Errorf("error getting nodes for %v: %v", nodePool.Description(cp.namespacePrefix), err)
	}

	var n []corev1.Node
	for _, node := range nodes {
		n = append(n, *node)
	}

	ready := kubernetes.NumReadyNodes(n)
	status := &latticev1.NodePoolStatusEpoch{
		NumInstances: ready,
		InstanceType: epochSpec.InstanceType,
		State:        latticev1.NodePoolStateScaling,
	}

	if ready == epochSpec.NumInstances {
		status.State = latticev1.NodePoolStateStable
	}

	return status, nil
}
