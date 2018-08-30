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
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	return record.nodePools, nil
}

func (b *NodePoolBackend) Get(path tree.PathSubcomponent) (*v1.NodePool, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	for _, nodePool := range record.nodePools {
		if nodePool.Path == path {
			np := &nodePool
			return np, nil
		}
	}

	return nil, nil
}
