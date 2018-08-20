package resolver

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type SecretStore interface {
	Ready() bool
	Put(systemID v1.SystemID, path tree.PathSubcomponent, v string) error
	Get(systemID v1.SystemID, path tree.PathSubcomponent) (string, error)
}

type SecretDoesNotExistError struct{}

func (e *SecretDoesNotExistError) Error() string {
	return "Secret does not exist"
}

func NewMemorySecretStore() SecretStore {
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

func (s *MemorySecretStore) Put(systemID v1.SystemID, path tree.PathSubcomponent, v string) error {
	s.store[s.keyString(systemID, path)] = v
	return nil
}

func (s *MemorySecretStore) Get(systemID v1.SystemID, path tree.PathSubcomponent) (string, error) {
	v, ok := s.store[s.keyString(systemID, path)]
	if !ok {
		return "", &SecretDoesNotExistError{}
	}

	return v, nil
}

func (s *MemorySecretStore) keyString(systemID v1.SystemID, path tree.PathSubcomponent) string {
	return fmt.Sprintf("%v.%v", systemID, path.String())
}
