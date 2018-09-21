package v1

import "time"

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

type System struct {
	ID SystemID `json:"id"`

	DefinitionURL string `json:"definitionUrl"`

	Status SystemStatus `json:"status"`
}

type SystemStatus struct {
	State SystemState `json:"state"`

	Version *Version `json:"version,omitempty"`

	CreationTimestamp time.Time  `json:"createdTimestamp"`
	DeletionTimestamp *time.Time `json:"deletionTimestamp,omitempty"`
}
