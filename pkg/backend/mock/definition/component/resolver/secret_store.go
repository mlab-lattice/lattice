package resolver

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func NewMemorySecretStore() *MemorySecretStore {
	return &MemorySecretStore{
		store: make(map[string]string),
	}
}

// MemorySecretStore implements a basic SecretStore that holds the Secrets in memory.
type MemorySecretStore struct {
	store map[string]string
}

func (s *MemorySecretStore) Ready() bool {
	return true
}

func (s *MemorySecretStore) Get(systemID v1.SystemID, path tree.PathSubcomponent) (string, error) {
	v, ok := s.store[s.keyString(systemID, path)]
	if !ok {
		return "", &resolver.SecretDoesNotExistError{}
	}

	return v, nil
}

func (s *MemorySecretStore) keyString(systemID v1.SystemID, path tree.PathSubcomponent) string {
	return fmt.Sprintf("%v.%v", systemID, path.String())
}
