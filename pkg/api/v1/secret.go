package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type Secret struct {
	// Secret service path
	Path tree.NodePath `json:"path"`
	// Name
	Name string `json:"name"`
	// Value
	Value string `json:"value"`
}
