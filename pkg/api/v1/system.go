package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type (
	SystemID      string
	SystemState   string
	SystemVersion string
)

const (
	SystemStatePending  SystemState = "pending"
	SystemStateFailed   SystemState = "failed"
	SystemStateDeleting SystemState = "deleting"

	SystemStateStable   SystemState = "stable"
	SystemStateDegraded SystemState = "degraded"
	SystemStateScaling  SystemState = "scaling"
	SystemStateUpdating SystemState = "updating"
)

type System struct {
	ID            SystemID              `json:"id"`
	State         SystemState           `json:"state"`
	DefinitionURL string                `json:"definitionUrl"`
	Services      map[tree.Path]Service `json:"services"`
}
