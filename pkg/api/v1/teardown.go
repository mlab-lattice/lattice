package v1

type (
	TeardownID    string
	TeardownState string
)

const (
	TeardownStatePending    TeardownState = "pending"
	TeardownStateInProgress TeardownState = "in progress"
	TeardownStateSucceeded  TeardownState = "succeeded"
	TeardownStateFailed     TeardownState = "failed"
)

type SystemTeardown struct {
	ID    TeardownID    `json:"id"`
	State TeardownState `json:"state"`
}
