package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/time"
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

	Successes int32 `json:"successes"`
	Failures  int32 `json:"failures"`

	StartTimestamp      *time.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`
}

type (
	JobRunID    string
	JobRunState string
)

const (
	JobRunStatePending   JobRunState = "pending"
	JobRunStateRunning   JobRunState = "running"
	JobRunStateSucceeded JobRunState = "succeeded"
	JobRunStateFailed    JobRunState = "failed"
	JobRunStateUnknown   JobRunState = "unknown"
)

type JobRun struct {
	ID JobRunID `json:"id"`

	Status JobRunStatus `json:"status"`
}

type JobRunStatus struct {
	State JobRunState `json:"state"`

	StartTimestamp      *time.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`
}
