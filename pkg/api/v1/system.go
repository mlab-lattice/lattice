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

// System API object documentation goes here
type System struct {
	// System ID
	ID SystemID `json:"id"`
	// State. One of (pending, failed, deleting, stable, degraded, scaling, updating)
	State SystemState `json:"state"`
	// git url for for where the definition lives in
	DefinitionURL string `json:"definitionUrl" example:"git://github.com/foo/foo.git"`
	// map for service path and services currently running in the system
	Services map[tree.NodePath]Service `json:"services"`
}
