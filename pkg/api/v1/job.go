package v1

import (
	"github.com/mlab-lattice/lattice/pkg/util/time"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type (
	JobID    string
	JobState string
)

const (
	JobStatePending  JobState = "pending"
	JobStateDeleting JobState = "deleting"

	JobStateQueued    JobState = "queued"
	JobStateRunning   JobState = "running"
	JobStateSucceeded JobState = "succeeded"
	JobStateFailed    JobState = "failed"
)

type Job struct {
	ID   JobID     `json:"id"`
	Path tree.Path `json:"path"`

	Status JobStatus `json:"status"`
}

type JobStatus struct {
	State JobState `json:"state"`

	Attempts map[string]JobStatusAttempt `json:"attempts,omitempty"`

	Successes int32 `json:"successes"`
	Failures  int32 `json:"failures"`

	StartTimestamp      *time.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`
}

type JobStatusAttempt struct {
}
