package v1

import "github.com/mlab-lattice/lattice/pkg/util/time"

type (
	SystemID    string
	SystemState string
	Version     string
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

// swagger:model System
type System struct {
	ID SystemID `json:"id"`

	DefinitionURL string `json:"definitionUrl"`

	Status SystemStatus `json:"status"`
}

type SystemStatus struct {
	State SystemState `json:"state"`

	Version *Version `json:"version,omitempty"`

	CreationTimestamp time.Time  `json:"creationTimestamp"`
	DeletionTimestamp *time.Time `json:"deletionTimestamp,omitempty"`
}
