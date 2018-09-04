package backend

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/mock/backend/system"
	"github.com/mlab-lattice/lattice/pkg/api/server/v1"
)

type MockBackend struct {
	systems *system.Backend
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		systems: system.NewBackend(),
	}
}

func (b *MockBackend) Systems() v1.SystemBackend {
	return b.systems
}
