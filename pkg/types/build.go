package types

import (
	// TODO: feels a little weird to have to import this here. should type definitions under pkg/system be moved into pkg/types?
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type BuildID string
type BuildState string

const (
	BuildStatePending   BuildState = "pending"
	BuildStateRunning   BuildState = "running"
	BuildStateSucceeded BuildState = "succeeded"
	BuildStateFailed    BuildState = "failed"
)

type SystemBuild struct {
	ID    BuildID    `json:"id"`
	State BuildState `json:"state"`

	Version SystemVersion `json:"version"`
	// Services maps service paths (e.g. /foo/bar/buzz) to the
	// ServiceBuild for that service in the SystemBuild.
	Services map[tree.NodePath]ServiceBuild `json:"serviceBuilds"`
}

type ServiceBuildID string
type ServiceBuildState string

const (
	ServiceBuildStatePending   ServiceBuildState = "pending"
	ServiceBuildStateRunning   ServiceBuildState = "running"
	ServiceBuildStateSucceeded ServiceBuildState = "succeeded"
	ServiceBuildStateFailed    ServiceBuildState = "failed"
)

type ServiceBuild struct {
	ID    ServiceBuildID    `json:"id"`
	State ServiceBuildState `json:"state"`

	// Components maps the component name to the build for that component.
	Components map[string]ComponentBuild `json:"componentBuilds"`
}

type ComponentBuildID string
type ComponentBuildState string
type ComponentBuildPhase string

const (
	ComponentBuildPhasePullingGitRepository ComponentBuildPhase = "pulling git repository"
	ComponentBuildPhaseBuildingDockerImage  ComponentBuildPhase = "building docker image"
	ComponentBuildPhasePushingDockerImage   ComponentBuildPhase = "pushing docker image"

	ComponentBuildStatePending   ComponentBuildState = "pending"
	ComponentBuildStateQueued    ComponentBuildState = "queued"
	ComponentBuildStateRunning   ComponentBuildState = "running"
	ComponentBuildStateSucceeded ComponentBuildState = "succeeded"
	ComponentBuildStateFailed    ComponentBuildState = "failed"
)

type ComponentBuild struct {
	ID                ComponentBuildID     `json:"id"`
	State             ComponentBuildState  `json:"state"`
	LastObservedPhase *ComponentBuildPhase `json:"lastObservedPhase,omitempty"`
	FailureMessage    *string              `json:"failureMessage,omitempty"`
}

type ComponentBuildFailureInfo struct {
	Message  string `json:"message"`
	Internal bool   `json:"internal"`
}
