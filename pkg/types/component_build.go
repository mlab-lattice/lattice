package types

type ComponentBuildID string
type ComponentBuildState string
type ComponentBuildPhase string

const (
	ComponentBuildPhasePullingGitRepository ComponentBuildPhase = "pulling git repository"
	ComponentBuildPhaseBuildingDockerImage  ComponentBuildPhase = "building docker image"
	ComponentBuildPhasePushingDockerImage   ComponentBuildPhase = "pushing docker image"

	ComponentBuildStatePending   ComponentBuildState = "Pending"
	ComponentBuildStateQueued    ComponentBuildState = "Queued"
	ComponentBuildStateRunning   ComponentBuildState = "Running"
	ComponentBuildStateSucceeded ComponentBuildState = "Succeeded"
	ComponentBuildStateFailed    ComponentBuildState = "Failed"
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
