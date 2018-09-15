package system

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type NodePoolBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

func (b *NodePoolBackend) List() ([]v1.NodePool, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	var nodePools []v1.NodePool
	for _, nodePool := range record.NodePools {
		nodePools = append(nodePools, *nodePool)
	}

	return nodePools, nil
}

func (b *NodePoolBackend) Get(path tree.PathSubcomponent) (*v1.NodePool, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	nodePool, ok := record.NodePools[path]
	if !ok {
		return nil, v1.NewInvalidPathError()
	}

	return nodePool, nil
}
