package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/time"
)

type (
	BuildID    string
	BuildState string
)

const (
	BuildStatePending  BuildState = "pending"
	BuildStateAccepted BuildState = "accepted"
	// FIXME(kevindrosendahl): should probably standardize on running vs in progress
	BuildStateRunning   BuildState = "running"
	BuildStateSucceeded BuildState = "succeeded"
	BuildStateFailed    BuildState = "failed"
)

// swagger:model Build
type Build struct {
	ID BuildID `json:"id"`

	Path    *tree.Path `json:"path,omitempty"`
	Version *Version   `json:"version,omitempty"`

	Status BuildStatus `json:"status"`
}

type BuildStatus struct {
	State   BuildState `json:"state"`
	Message string     `json:"message,omitempty"`

	StartTimestamp      *time.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`

	Path    *tree.Path `json:"path,omitempty"`
	Version *Version   `json:"version,omitempty"`

	// Workloads maps component paths (e.g. /foo/bar/buzz) to the
	// status of the build for that service in the Build.
	Workloads map[tree.Path]WorkloadBuild `json:"workloads"`
}

type WorkloadBuild struct {
	ContainerBuild
	Sidecars map[string]ContainerBuild `json:"sidecars,omitempty"`
}

type (
	ContainerBuildState string
	ContainerBuildPhase string
	ContainerBuildID    string
)

const (
	ContainerBuildPhasePullingGitRepository ContainerBuildPhase = "pulling git repository"
	ContainerBuildPhasePullingDockerImage   ContainerBuildPhase = "pulling docker image"
	ContainerBuildPhaseBuildingDockerImage  ContainerBuildPhase = "building docker image"
	ContainerBuildPhasePushingDockerImage   ContainerBuildPhase = "pushing docker image"

	ContainerBuildStatePending   ContainerBuildState = "pending"
	ContainerBuildStateQueued    ContainerBuildState = "queued"
	ContainerBuildStateRunning   ContainerBuildState = "running"
	ContainerBuildStateSucceeded ContainerBuildState = "succeeded"
	ContainerBuildStateFailed    ContainerBuildState = "failed"
)

type ContainerBuild struct {
	ID ContainerBuildID `json:"id"`

	Status ContainerBuildStatus `json:"status"`
}

type ContainerBuildStatus struct {
	State ContainerBuildState `json:"state"`

	StartTimestamp      *time.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`

	LastObservedPhase *ContainerBuildPhase `json:"lastObservedPhase,omitempty"`
	FailureMessage    *string              `json:"failureMessage,omitempty"`
}

type ContainerBuildFailureInfo struct {
	Message  string `json:"message"`
	Internal bool   `json:"internal"`
}
