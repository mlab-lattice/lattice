package types

type ComponentBuildID string
type ComponentBuildState string
type ComponentBuildPhase string

type ComponentBuild struct {
	ID                ComponentBuildID     `json:"id"`
	State             ComponentBuildState  `json:"state"`
	LastObservedPhase *ComponentBuildPhase `json:"lastObservedPhase,omitempty"`
	FailureMessage    *string              `json:"failureMessage,omitempty"`
}
