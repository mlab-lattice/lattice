package types

import (
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type Service struct {
	ID      string        `json:"id"`
	Path    tree.NodePath `json:"path"`
	Address *string       `json:"address"`
}
