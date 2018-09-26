package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	BuildKind     = SchemeGroupVersion.WithKind("Build")
	BuildListKind = SchemeGroupVersion.WithKind("BuildList")

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

func (b *Build) DefinitionVersionLabel() (v1.Version, bool) {
	version, ok := b.Labels[BuildDefinitionVersionLabelKey]
	return v1.Version(version), ok
}

func (b *Build) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, b.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", b.Namespace))
	}

	version := v1.Version("unknown")
	if label, ok := b.DefinitionVersionLabel(); ok {
		version = label
	}

	return fmt.Sprintf("build %v (version %v in system %v)", b.Name, version, systemID)
}

type BuildSpec struct {
	Version *v1.Version `json:"version"`
	Path    *tree.Path  `json:"path"`
}

type BuildStatus struct {
	// Build specs are immutable so no need for ObservedGeneration

	State   BuildState `json:"state"`
	Message string     `json:"message"`

	InternalError *string `json:"internalError,omitempty"`

	Definition *resolver.ResolutionTree `json:"definition,omitempty"`
	Path       *tree.Path               `json:"path,omitempty"`
	Version    *v1.Version              `json:"version,omitempty"`

	StartTimestamp      *metav1.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`

	// Maps a workload path to the information about its container builds
	Workloads map[tree.Path]BuildStatusWorkload `json:"workloads"`

	// Maps a container build's ID to its status
	ContainerBuildStatuses map[v1.ContainerBuildID]ContainerBuildStatus `json:"containerBuildStatuses"`
}

type BuildStatusWorkload struct {
	MainContainer v1.ContainerBuildID            `json:"mainContainer"`
	Sidecars      map[string]v1.ContainerBuildID `json:"sidecars"`
}

type BuildState string

const (
	BuildStatePending   BuildState = ""
	BuildStateAccepted  BuildState = "accepted"
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
