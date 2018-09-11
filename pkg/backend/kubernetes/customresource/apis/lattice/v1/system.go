package v1

import (
	"fmt"

	"encoding/json"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularSystem = "system"
	ResourcePluralSystem   = "systems"
	ResourceScopeSystem    = apiextensionsv1beta1.NamespaceScoped
)

var (
	SystemKind     = SchemeGroupVersion.WithKind("System")
	SystemListKind = SchemeGroupVersion.WithKind("SystemList")

	SystemDefinitionVersionLabelKey = fmt.Sprintf("system.%v/definition-version", GroupName)
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

func (s *System) ResourceNamespace(namespacePrefix string) string {
	return kubeutil.SystemNamespace(namespacePrefix, s.V1ID())
}

func (s *System) Stable() bool {
	return s.UpdateProcessed() && s.Status.State == SystemStateStable
}

func (s *System) UpdateProcessed() bool {
	return s.Status.ObservedGeneration >= s.Generation
}

func (s *System) Description() string {
	return fmt.Sprintf("system %v", s.V1ID())
}

// N.B.: important: if you update the SystemSpec or SystemSpecServiceInfo you must also update
// the systemSpecEncoder and SystemSpec's UnmarshalJSON
// +k8s:deepcopy-gen=false
type SystemSpec struct {
	DefinitionURL string `json:"definitionUrl"`

	Definition             *resolver.ComponentTree           `json:"definition"`
	WorkloadBuildArtifacts *SystemSpecWorkloadBuildArtifacts `json:"workloadBuildArtifacts"`
}

// +k8s:deepcopy-gen=false
type SystemSpecWorkloadBuildArtifacts struct {
	inner *tree.JSONRadix
}

func NewSystemSpecWorkloadBuildArtifacts() *SystemSpecWorkloadBuildArtifacts {
	return &SystemSpecWorkloadBuildArtifacts{
		inner: tree.NewJSONRadix(
			func(i interface{}) (json.RawMessage, error) {
				return json.Marshal(&i)
			},
			func(data json.RawMessage) (interface{}, error) {
				var w WorkloadContainerBuildArtifacts
				if err := json.Unmarshal(data, &w); err != nil {
					return nil, err
				}

				return w, nil
			},
		),
	}
}

func (a *SystemSpecWorkloadBuildArtifacts) Insert(
	p tree.Path,
	w WorkloadContainerBuildArtifacts,
) (WorkloadContainerBuildArtifacts, bool) {
	i, ok := a.inner.Insert(p, w)
	if !ok {
		return WorkloadContainerBuildArtifacts{}, false
	}

	return i.(WorkloadContainerBuildArtifacts), true
}

func (a *SystemSpecWorkloadBuildArtifacts) Get(p tree.Path) (WorkloadContainerBuildArtifacts, bool) {
	i, ok := a.inner.Get(p)
	if !ok {
		return WorkloadContainerBuildArtifacts{}, false
	}

	return i.(WorkloadContainerBuildArtifacts), true
}

func (a *SystemSpecWorkloadBuildArtifacts) ReplacePrefix(p tree.Path, other *SystemSpecWorkloadBuildArtifacts) {
	a.inner.ReplacePrefix(p, other.inner.Radix)
}

func (a *SystemSpecWorkloadBuildArtifacts) MarshalJSON() ([]byte, error) {
	return json.Marshal(&a.inner)
}

func (a *SystemSpecWorkloadBuildArtifacts) UnmarshalJSON(data []byte) error {
	a2 := NewSystemSpecWorkloadBuildArtifacts()
	if err := json.Unmarshal(data, &a2.inner); err != nil {
		return err
	}

	*a = *a2
	return nil
}

// +k8s:deepcopy-gen=false
type SystemStatus struct {
	ObservedGeneration int64 `json:"observedGeneration"`

	State SystemState `json:"state"`

	Services  map[tree.Path]SystemStatusService              `json:"services"`
	NodePools map[tree.PathSubcomponent]SystemStatusNodePool `json:"nodePools"`
}

type SystemStatusService struct {
	Name       string `json:"name"`
	Generation int64  `json:"generation"`
	ServiceStatus
}

type SystemStatusNodePool struct {
	Name       string `json:"name"`
	Generation int64  `json:"generation"`
	NodePoolStatus
}

type SystemState string

const (
	// lifecycle states
	// note that the "deleting" state is implied by a non-nil metadata.DeletionTimestamp
	SystemStatePending SystemState = ""
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
