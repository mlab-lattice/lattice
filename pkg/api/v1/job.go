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
	// ID
	ID JobID `json:"id"`
	// Path
	Path tree.NodePath `json:"path"`
	// State
	State JobState `json:"state"`
	// StartTimestamp
	StartTimestamp *time.Time `json:"startTimestamp,omitempty"`
	// CompletionTimestamp
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`
}
