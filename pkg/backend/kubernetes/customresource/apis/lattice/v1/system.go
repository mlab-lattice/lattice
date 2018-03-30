package v1

import (
	"encoding/json"

	"github.com/mlab-lattice/system/pkg/api/v1"
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

func (s *System) V1ID() v1.SystemID {
	return v1.SystemID(s.Name)
}

// N.B.: important: if you update the SystemSpec or SystemSpecServiceInfo you must also update
// the systemSpecEncoder and SystemSpec's UnmarshalJSON
// +k8s:deepcopy-gen=false
type SystemSpec struct {
	DefinitionURL string                                  `json:"definitionUrl"`
	Services      map[tree.NodePath]SystemSpecServiceInfo `json:"services"`
}

// +k8s:deepcopy-gen=false
type SystemSpecServiceInfo struct {
	Definition definition.Service `json:"definition"`

	// ComponentBuildArtifacts maps Component names to the artifacts created by their build
	ComponentBuildArtifacts map[string]ComponentBuildArtifacts `json:"componentBuildArtifacts"`
}

type systemSpecServiceInfoEncoder struct {
	Definition              json.RawMessage                    `json:"definition"`
	ComponentBuildArtifacts map[string]ComponentBuildArtifacts `json:"componentBuildArtifacts"`
}

func (i *SystemSpecServiceInfo) UnmarshalJSON(data []byte) error {
	var decoded systemSpecServiceInfoEncoder
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	service, err := definition.NewServiceFromJSON(decoded.Definition)
	if err != nil {
		return err
	}

	*i = SystemSpecServiceInfo{
		Definition:              service,
		ComponentBuildArtifacts: decoded.ComponentBuildArtifacts,
	}
	return nil
}

// +k8s:deepcopy-gen=false
type SystemStatus struct {
	State              SystemState `json:"state"`
	ObservedGeneration int64       `json:"observedGeneration"`

	// FIXME: remove this when ObservedGeneration is supported for CRD
	UpdateProcessed bool `json:"updateProcessed"`

	// Maps a Service path to its Service.Status
	Services map[tree.NodePath]SystemStatusService `json:"services"`
}

type SystemStatusService struct {
	Name string `json:"name"`
	ServiceStatus
}

type SystemState string

const (
	// lifecycle states
	SystemStatePending SystemState = "pending"
	SystemStateFailed  SystemState = "failed"

	// transient states once the system has been created
	SystemStateStable   SystemState = "stable"
	SystemStateDegraded SystemState = "degraded"
	SystemStateScaling  SystemState = "scaling"
	SystemStateUpdating SystemState = "updating"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []System `json:"items"`
}
