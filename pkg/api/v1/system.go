package v1

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
	ID            SystemID    `json:"id"`
	State         SystemState `json:"state"`
	DefinitionURL string      `json:"definitionUrl"`
}
