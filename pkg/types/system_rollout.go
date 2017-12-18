package types

type SystemRolloutID string
type SystemRolloutState string

const (
	SystemRolloutStatePending    SystemRolloutState = "Pending"
	SystemRolloutStateAccepted   SystemRolloutState = "Accepted"
	SystemRolloutStateInProgress SystemRolloutState = "InProgress"
	SystemRolloutStateSucceeded  SystemRolloutState = "Succeeded"
	SystemRolloutStateFailed     SystemRolloutState = "Failed"
)

type SystemRollout struct {
	ID      SystemRolloutID    `json:"id"`
	BuildID SystemBuildID      `json:"buildId"`
	State   SystemRolloutState `json:"state"`
}
