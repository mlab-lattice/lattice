package v1

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	"github.com/mlab-lattice/system/pkg/types"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularServiceBuild  = "servicebuild"
	ResourcePluralServiceBuild    = "servicebuilds"
	ResourceShortNameServiceBuild = "lsvcb"
	ResourceScopeServiceBuild     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ServiceBuildSpec   `json:"spec"`
	Status            ServiceBuildStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=false
type ServiceBuildSpec struct {
	Components map[string]ServiceBuildComponentBuildInfo `json:"components"`
}

// +k8s:deepcopy-gen=false
type ServiceBuildComponentBuildInfo struct {
	DefinitionBlock   block.ComponentBuild       `json:"definitionBlock"`
	DefinitionHash    *string                    `json:"definitionHash,omitempty"`
	BuildName         *string                    `json:"buildName,omitempty"`
	BuildState        *ComponentBuildState       `json:"buildState"`
	LastObservedPhase *types.ComponentBuildPhase `json:"lastObservedPhase,omitempty"`
	FailureInfo       *ComponentBuildFailureInfo `json:"failureInfo,omitempty"`
}

type ServiceBuildStatus struct {
	State   ServiceBuildState `json:"state,omitempty"`
	Message string            `json:"message,omitempty"`
}

type ServiceBuildState string

const (
	ServiceBuildStatePending   ServiceBuildState = "pending"
	ServiceBuildStateRunning   ServiceBuildState = "running"
	ServiceBuildStateSucceeded ServiceBuildState = "succeeded"
	ServiceBuildStateFailed    ServiceBuildState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ServiceBuild `json:"items"`
}
