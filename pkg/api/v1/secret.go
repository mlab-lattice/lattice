package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type Secret struct {
	Path  tree.PathSubcomponent `json:"path"`
	Value string                `json:"value"`
}
