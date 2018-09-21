package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"time"
)

type (
	DeployID    string
	DeployState string
)

const (
	DeployStatePending    DeployState = "pending"
	DeployStateAccepted   DeployState = "accepted"
	DeployStateInProgress DeployState = "in progress"
	DeployStateSucceeded  DeployState = "succeeded"
	DeployStateFailed     DeployState = "failed"
)

type Deploy struct {
	ID DeployID `json:"id"`

	Build   *BuildID   `json:"build,omitempty"`
	Path    *tree.Path `json:"path,omitempty"`
	Version *Version   `json:"version,omitempty"`

	Status DeployStatus `json:"status"`
}

type DeployStatus struct {
	State   DeployState `json:"state"`
	Message string      `json:"message,omitempty"`

	Build   *BuildID   `json:"build,omitempty"`
	Path    *tree.Path `json:"path,omitempty"`
	Version *Version   `json:"version,omitempty"`

	StartTimestamp      *time.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`
}
