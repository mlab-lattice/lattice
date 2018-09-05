package v1

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/v1/system"
)

type Backend struct {
	systems *system.Backend
}

func NewBackend() *Backend {
	return &Backend{
		systems: system.NewBackend(),
	}
}

func (b *Backend) Systems() v1.SystemBackend {
	return b.systems
}
