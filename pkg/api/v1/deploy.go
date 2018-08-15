package v1

type (
	DeployID    string
	DeployState string
)

const (
	DeployStatePending    DeployState = "pending"
	DeployStateAccepted   DeployState = "accepted"
	DeployStateInProgress DeployState = "in progress"
	DeployStateSucceeded  DeployState = "succeeded"
	DeployStateFailed     DeployState = "failed"
)

type Deploy struct {
	// ID
	ID DeployID `json:"id"`
	// Build ID
	BuildID BuildID `json:"buildId"`
	// State
	State DeployState `json:"state"`
}
