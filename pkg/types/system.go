package types

import (
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type SystemID string
type SystemVersion string

type System struct {
	ID            SystemID                  `json:"id"`
	DefinitionURL string                    `json:"definitionUrl"`
	Services      map[tree.NodePath]Service `json:"services"`
}
