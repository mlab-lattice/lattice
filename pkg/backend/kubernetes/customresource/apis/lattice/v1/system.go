package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

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

	NodePools map[string]NodePoolSpec             `json:"nodePools"`
	Services  map[tree.Path]SystemSpecServiceInfo `json:"services"`
	Jobs      map[tree.Path]SystemSpecJobInfo     `json:"jobs"`
}

// +k8s:deepcopy-gen=false
type SystemSpecServiceInfo struct {
	Definition *definitionv1.Service `json:"definition"`

	// ContainerBuildArtifacts maps container names to the artifacts created by their build
	ContainerBuildArtifacts map[string]ContainerBuildArtifacts `json:"containerBuildArtifacts"`
}

// +k8s:deepcopy-gen=false
type SystemSpecJobInfo struct {
	Definition *definitionv1.Job `json:"definition"`

	// ContainerBuildArtifacts maps container names to the artifacts created by their build
	ContainerBuildArtifacts map[string]ContainerBuildArtifacts `json:"containerBuildArtifacts"`
}

// +k8s:deepcopy-gen=false
type SystemStatus struct {
	ObservedGeneration int64 `json:"observedGeneration"`

	State SystemState `json:"state"`

	Services  map[tree.Path]SystemStatusService `json:"services"`
	NodePools map[string]SystemStatusNodePool   `json:"nodePools"`
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
