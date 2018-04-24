package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularDeploy = "deploy"
	ResourcePluralDeploy   = "deploys"
	ResourceScopeDeploy    = apiextensionsv1beta1.NamespaceScoped
)

var (
	DeployKind                      = SchemeGroupVersion.WithKind("Deploy")
	DeployDefinitionURLLabelKey     = fmt.Sprintf("deploy.%v/definition-url", GroupName)
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

func (d *Deploy) BuildIDLabel() (string, bool) {
	id, ok := d.Labels[BuildIDLabelKey]
	return id, ok
}

func (d *Deploy) DefinitionURLLabel() (string, bool) {
	url, ok := d.Labels[DeployDefinitionURLLabelKey]
	return url, ok
}

func (d *Deploy) DefinitionVersionLabel() (string, bool) {
	version, ok := d.Labels[DeployDefinitionVersionLabelKey]
	return version, ok
}

func (d *Deploy) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, d.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", d.Namespace))
	}

	version := "unknown"
	if label, ok := d.DefinitionVersionLabel(); ok {
		version = label
	}

	definitionURL := "unknown definition URL"
	if label, ok := d.DefinitionURLLabel(); ok {
		definitionURL = label
	}

	return fmt.Sprintf("deploy %v (build %v, version %v of %v in system %v)", d.Name, d.Spec.BuildName, version, definitionURL, systemID)
}

type DeploySpec struct {
	BuildName string `json:"buildName"`
}

type DeployStatus struct {
	// Deploys are immutable so no need for ObservedGeneration

	State   DeployState `json:"state"`
	Message string      `json:"message"`

	StartTimestamp      *metav1.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`
}

type DeployState string

const (
	DeployStatePending    DeployState = "pending"
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
