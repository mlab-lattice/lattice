package v1

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	"github.com/mlab-lattice/system/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularComponentBuild  = "componentbuild"
	ResourcePluralComponentBuild    = "componentbuilds"
	ResourceShortNameComponentBuild = "lcb"
	ResourceScopeComponentBuild     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ComponentBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ComponentBuildSpec   `json:"spec"`
	Status            ComponentBuildStatus `json:"status"`
}

// +k8s:deepcopy-gen=false
type ComponentBuildSpec struct {
	BuildDefinitionBlock block.ComponentBuild     `json:"definitionBlock"`
	Artifacts            *ComponentBuildArtifacts `json:"artifacts,omitempty"`
}

// +k8s:deepcopy-gen=false
type ComponentBuildArtifacts struct {
	DockerImageFqn string `json:"dockerImageFqn"`
}

type ComponentBuildStatus struct {
	State             ComponentBuildState        `json:"state"`
	LastObservedPhase *types.ComponentBuildPhase `json:"lastObservedPhase,omitempty"`
	FailureInfo       *ComponentBuildFailureInfo `json:"failureInfo,omitempty"`
}

type ComponentBuildState string

const (
	ComponentBuildStatePending   ComponentBuildState = "pending"
	ComponentBuildStateQueued    ComponentBuildState = "queued"
	ComponentBuildStateRunning   ComponentBuildState = "running"
	ComponentBuildStateSucceeded ComponentBuildState = "succeeded"
	ComponentBuildStateFailed    ComponentBuildState = "failed"
)

type ComponentBuildFailureInfo struct {
	Message  string `json:"message"`
	Internal bool   `json:"internal"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ComponentBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ComponentBuild `json:"items"`
}
