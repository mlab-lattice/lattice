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

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	var nodePools []v1.NodePool
	for _, nodePool := range record.nodePools {
		nodePools = append(nodePools, *nodePool)
	}

	return nodePools, nil
}

func (b *NodePoolBackend) Get(path tree.PathSubcomponent) (*v1.NodePool, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	for _, nodePool := range record.nodePools {
		if nodePool.Path == path {
			result := new(v1.NodePool)
			*result = *nodePool
			return result, nil
		}
	}

	return nil, v1.NewInvalidNodePoolPathError()
}
