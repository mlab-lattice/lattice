package v1

import (
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type Secret struct {
	Path  tree.NodePath `json:"path"`
	Name  string        `json:"name"`
	Value string        `json:"value"`
}
