package types

type SystemTeardownID string
type SystemTeardownState string

const (
	SystemTeardownStatePending    SystemTeardownState = "Pending"
	SystemTeardownStateInProgress SystemTeardownState = "InProgress"
	SystemTeardownStateSucceeded  SystemTeardownState = "Succeeded"
	SystemTeardownStateFailed     SystemTeardownState = "Failed"
)

type SystemTeardown struct {
	ID    SystemTeardownID    `json:"id"`
	State SystemTeardownState `json:"state"`
}
