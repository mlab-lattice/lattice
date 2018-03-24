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
	ID      DeployID    `json:"id"`
	BuildID BuildID     `json:"buildId"`
	State   DeployState `json:"state"`
}
