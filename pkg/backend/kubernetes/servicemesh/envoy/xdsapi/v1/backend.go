package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type Backend interface {
	Ready() bool
	Services(serviceCluster string) (map[tree.Path]*Service, error)
}
