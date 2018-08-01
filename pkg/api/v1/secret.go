package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type Secret struct {
	Path  tree.Path `json:"path"`
	Name  string    `json:"name"`
	Value string    `json:"value"`
}
