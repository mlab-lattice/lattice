package mock

import (
	"github.com/mlab-lattice/lattice/pkg/api/server/mock/backend/system"
)

type MockBackend struct {
	systems *system.Backend
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		systems: system.NewBackend(),
	}
}

func (b *MockBackend) Systems() *system.Backend {
	return b.systems
}
