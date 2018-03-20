package types

import (
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type SystemID string
type SystemState string
type SystemVersion string

const (
	SystemStateScaling  SystemState = "scaling"
	SystemStateUpdating SystemState = "updating"
	SystemStateStable   SystemState = "stable"
	SystemStateFailed   SystemState = "failed"
	SystemStateDeleting SystemState = "deleting"
)

type System struct {
	ID            SystemID                  `json:"id"`
	State         SystemState               `json:"state"`
	DefinitionURL string                    `json:"definitionUrl"`
	Services      map[tree.NodePath]Service `json:"services"`
}
