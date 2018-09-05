package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularDeploy = "deploy"
	ResourcePluralDeploy   = "deploys"
	ResourceScopeDeploy    = apiextensionsv1beta1.NamespaceScoped
)

var (
	DeployKind     = SchemeGroupVersion.WithKind("Deploy")
	DeployListKind = SchemeGroupVersion.WithKind("DeployList")

	DeployIDLabelKey                = fmt.Sprintf("deploy.%v/id", GroupName)
	DeployDefinitionVersionLabelKey = fmt.Sprintf("deploy.%v/definition-version", GroupName)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Deploy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DeploySpec   `json:"spec"`
	Status            DeployStatus `json:"status"`
}

func (d *Deploy) BuildIDLabel() (v1.BuildID, bool) {
	id, ok := d.Labels[BuildIDLabelKey]
	return v1.BuildID(id), ok
}

func (d *Deploy) DefinitionVersionLabel() (v1.SystemVersion, bool) {
	version, ok := d.Labels[DeployDefinitionVersionLabelKey]
	return v1.SystemVersion(version), ok
}

func (d *Deploy) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, d.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", d.Namespace))
	}

	version := v1.SystemVersion("unknown")
	if label, ok := d.DefinitionVersionLabel(); ok {
		version = label
	}

	buildID := v1.BuildID("unknown")
	if label, ok := d.BuildIDLabel(); ok {
		buildID = label
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
	Version *DeploySpecVersionInfo
}

type DeploySpecVersionInfo struct {
	Version v1.SystemVersion
	Path    tree.Path
}

type DeployStatus struct {
	// Deploy specs are immutable so no need for ObservedGeneration

	State   DeployState `json:"state"`
	Message string      `json:"message,omitempty"`

	BuildID *v1.BuildID `json:"buildId,omitempty"`

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
