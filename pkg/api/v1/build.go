package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"time"
)

type (
	BuildID    string
	BuildState string
)

const (
	BuildStatePending   BuildState = "pending"
	BuildStateAccepted  BuildState = "accepted"
	BuildStateRunning   BuildState = "running"
	BuildStateSucceeded BuildState = "succeeded"
	BuildStateFailed    BuildState = "failed"
)

type Build struct {
	ID BuildID `json:"id"`

	State   BuildState `json:"state"`
	Message string     `json:"message,omitempty"`

	StartTimestamp      *time.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`

	Version *Version   `json:"version,omitempty"`
	Path    *tree.Path `json:"path,omitempty"`

	// Components maps component paths (e.g. /foo/bar/buzz) to the
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
	ID    ContainerBuildID    `json:"id"`
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
