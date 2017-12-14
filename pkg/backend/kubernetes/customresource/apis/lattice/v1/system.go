package v1

import (
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularSystem  = "system"
	ResourcePluralSystem    = "systems"
	ResourceShortNameSystem = "lsys"
	ResourceScopeSystem     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type System struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SystemSpec   `json:"spec"`
	Status            SystemStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen=false
type SystemSpec struct {
	Services map[tree.NodePath]SystemSpecServiceInfo `json:"services"`
}

// +k8s:deepcopy-gen=false
type SystemSpecServiceInfo struct {
	Definition definition.Service `json:"definition"`

	// ComponentBuildArtifacts maps Component names to the artifacts created by their build
	ComponentBuildArtifacts map[string]ComponentBuildArtifacts `json:"componentBuildArtifacts"`
}

type SystemStatus struct {
	State              SystemState                               `json:"state"`
	ObservedGeneration int64                                     `json:"observedGeneration"`
	Services           map[tree.NodePath]SystemStatusServiceInfo `json:"services"`
}

type SystemState string

const (
	SystemStateScaling  SystemState = "scaling"
	SystemStateUpdating SystemState = "updating"
	SystemStateStable   SystemState = "stable"
	SystemStateFailed   SystemState = "failed"
)

type SystemStatusServiceInfo struct {
	Name   string        `json:"name"`
	Status ServiceStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []System `json:"items"`
}
