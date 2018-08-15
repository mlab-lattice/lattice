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

type Teardown struct {
	// ID
	ID TeardownID `json:"id"`
	// State
	State TeardownState `json:"state"`
}
