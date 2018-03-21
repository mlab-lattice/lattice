package types

import (
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type SystemID string
type SystemState string
type SystemVersion string

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
	ID            SystemID                  `json:"id"`
	State         SystemState               `json:"state"`
	DefinitionURL string                    `json:"definitionUrl"`
	Services      map[tree.NodePath]Service `json:"services"`
}
