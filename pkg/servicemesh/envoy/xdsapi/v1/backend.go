package v1

import (
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type Backend interface {
	Ready() bool
	Services(serviceCluster string) (map[tree.NodePath]*Service, error)
}
