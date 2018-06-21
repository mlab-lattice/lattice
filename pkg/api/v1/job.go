package v1

import (
	"time"

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
	ID    JobID         `json:"id"`
	Path  tree.NodePath `json:"path"`
	State JobState      `json:"state"`

	StartTimestamp      *time.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`
}
