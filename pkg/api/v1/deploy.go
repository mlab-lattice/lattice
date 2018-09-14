package v1

import "github.com/mlab-lattice/lattice/pkg/definition/tree"

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
	ID      DeployID    `json:"id"`
	Build   *BuildID    `json:"build,omitempty"`
	Path    *tree.Path  `json:"path,omitempty"`
	Version *Version    `json:"version,omitempty"`
	State   DeployState `json:"state"`
	Message string      `json:"message,omitempty"`
}
