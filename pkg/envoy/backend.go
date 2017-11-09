package envoy

import (
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"
)

type Backend interface {
	Ready() bool
	Services() (map[systemtree.NodePath]*Service, error)
}
