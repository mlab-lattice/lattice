package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

// swagger:model Secret
type Secret struct {
	Path  tree.PathSubcomponent `json:"path"`
	Value string                `json:"value"`
}
