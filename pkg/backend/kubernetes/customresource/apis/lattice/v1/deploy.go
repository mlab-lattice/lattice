package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DeployKind     = SchemeGroupVersion.WithKind("Deploy")
	DeployListKind = SchemeGroupVersion.WithKind("DeployList")
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Deploy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DeploySpec   `json:"spec"`
	Status            DeployStatus `json:"status"`
}

func (d *Deploy) V1ID() v1.DeployID {
	return v1.DeployID(d.Name)
}

func (d *Deploy) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, d.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", d.Namespace))
	}

	version := v1.Version("unknown")
	if d.Status.Version != nil {
		version = *d.Status.Version
	}

	buildID := v1.BuildID("unknown")
	if d.Status.Build != nil {
		buildID = *d.Status.Build
	}

	return fmt.Sprintf(
		"deploy %v (build %v, version %v (build %v) in system %v)",
		d.Name,
		d.Spec.Build,
		version,
		buildID,
		systemID,
	)
}

type DeploySpec struct {
	Build   *v1.BuildID `json:"build,omitempty"`
	Version *v1.Version `json:"version,omitempty"`
	Path    *tree.Path  `json:"path"`
}

type DeployStatus struct {
	// Deploy specs are immutable so no need for ObservedGeneration

	State   DeployState `json:"state"`
	Message string      `json:"message,omitempty"`

	InternalError *string `json:"internalError,omitempty"`

	Build   *v1.BuildID `json:"build,omitempty"`
	Path    *tree.Path  `json:"path,omitempty"`
	Version *v1.Version `json:"version,omitempty"`

	StartTimestamp      *metav1.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`
}

type DeployState string

const (
	DeployStatePending    DeployState = ""
	DeployStateAccepted   DeployState = "accepted"
	DeployStateInProgress DeployState = "in progress"
	DeployStateSucceeded  DeployState = "succeeded"
	DeployStateFailed     DeployState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DeployList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Deploy `json:"items"`
}
