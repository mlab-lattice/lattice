package types

type SystemRolloutID string
type SystemRolloutState string

type SystemRollout struct {
	ID      SystemRolloutID    `json:"id"`
	BuildID SystemBuildID      `json:"buildId"`
	State   SystemRolloutState `json:"state"`
}
