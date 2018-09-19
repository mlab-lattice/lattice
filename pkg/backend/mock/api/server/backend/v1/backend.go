package v1

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/v1/system"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
)

type Backend struct {
	systems *system.Backend
}

func NewBackend(componentResolver resolver.ComponentResolver) *Backend {
	return &Backend{system.NewBackend(componentResolver)}
}

func (b *Backend) Systems() v1.SystemBackend {
	return b.systems
}
