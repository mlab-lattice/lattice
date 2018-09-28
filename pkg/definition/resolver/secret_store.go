package resolver

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type SecretStore interface {
	Ready() bool
	Get(systemID v1.SystemID, path tree.PathSubcomponent) (string, error)
}

type SecretDoesNotExistError struct{}

func (e *SecretDoesNotExistError) Error() string {
	return "Secret does not exist"
}
