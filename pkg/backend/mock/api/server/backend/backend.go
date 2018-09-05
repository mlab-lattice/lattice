package backend

import (
	serverv1 "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	mockv1 "github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/v1"
)

func NewMockBackend() *MockBackend {
	return &MockBackend{
		v1: mockv1.NewBackend(),
	}
}

type MockBackend struct {
	v1 *mockv1.Backend
}

func (b *MockBackend) V1() serverv1.Backend {
	return b.v1
}
