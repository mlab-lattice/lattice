package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularBuild = "build"
	ResourcePluralBuild   = "builds"
	ResourceScopeBuild    = apiextensionsv1beta1.NamespaceScoped
)

var (
	BuildKind     = SchemeGroupVersion.WithKind("Build")
	BuildListKind = SchemeGroupVersion.WithKind("BuildList")

	BuildIDLabelKey                = fmt.Sprintf("build.%v/id", GroupName)
	BuildDefinitionVersionLabelKey = fmt.Sprintf("build.%v/definition-version", GroupName)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Build struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              BuildSpec   `json:"spec"`
	Status            BuildStatus `json:"status,omitempty"`
}

func (b *Build) DefinitionVersionLabel() (v1.SystemVersion, bool) {
	version, ok := b.Labels[BuildDefinitionVersionLabelKey]
	return v1.SystemVersion(version), ok
}

func (b *Build) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, b.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", b.Namespace))
	}

	version := v1.SystemVersion("unknown")
	if label, ok := b.DefinitionVersionLabel(); ok {
		version = label
	}

	return fmt.Sprintf("build %v (version %v in system %v)", b.Name, version, systemID)
}

// +k8s:deepcopy-gen=false
type BuildSpec struct {
	Version v1.SystemVersion `json:"version"`
}

type BuildStatus struct {
	// Build specs are immutable so no need for ObservedGeneration

	State   BuildState `json:"state"`
	Message string     `json:"message"`

	Definition     *definitionv1.SystemNode `json:"definition"`
	ResolutionInfo resolver.ResolutionInfo  `json:"resolutionInfo"`

	StartTimestamp      *metav1.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`

	// Maps a service path to the information about its container builds
	Services map[tree.Path]BuildStatusService `json:"services"`

	// Maps a service path to the information about its container builds
	Jobs map[tree.Path]BuildStatusJob `json:"jobs"`

	// Maps a ServiceBuild.Name to the ServiceBuild.Status
	ContainerBuildStatuses map[string]ContainerBuildStatus `json:"containerBuildStatuses"`
}

type BuildStatusService struct {
	MainContainer string            `json:"mainContainer"`
	Sidecars      map[string]string `json:"sidecars"`
}

type BuildStatusJob struct {
	MainContainer string            `json:"mainContainer"`
	Sidecars      map[string]string `json:"sidecars"`
}

type BuildState string

const (
	BuildStatePending   BuildState = ""
	BuildStateRunning   BuildState = "running"
	BuildStateSucceeded BuildState = "succeeded"
	BuildStateFailed    BuildState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Build `json:"items"`
}
