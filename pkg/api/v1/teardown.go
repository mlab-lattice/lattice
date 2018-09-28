package v1

import (
	"github.com/mlab-lattice/lattice/pkg/util/time"
)

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
	ID TeardownID `json:"id"`

	Status TeardownStatus `json:"status"`
}

type TeardownStatus struct {
	State   TeardownState `json:"state"`
	Message string        `json:"message,omitempty"`

	StartTimestamp      *time.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`
}
