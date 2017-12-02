package types

type SystemTeardownID string
type SystemTeardownState string

type SystemTeardown struct {
	ID    SystemTeardownID    `json:"id"`
	State SystemTeardownState `json:"state"`
}
