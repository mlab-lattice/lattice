package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/block"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularServiceBuild = "servicebuild"
	ResourcePluralServiceBuild   = "servicebuilds"
	ResourceScopeServiceBuild    = apiextensionsv1beta1.NamespaceScoped
)

var (
	ServiceBuildKind     = SchemeGroupVersion.WithKind("ServiceBuild")
	ServiceBuildListKind = SchemeGroupVersion.WithKind("ServiceBuildList")

	ServiceBuildDefinitionURLLabelKey     = fmt.Sprintf("service.build.%v/definition-url", GroupName)
	ServiceBuildDefinitionVersionLabelKey = fmt.Sprintf("service.build.%v/definition-version", GroupName)
	ServiceBuildPathLabelKey              = fmt.Sprintf("service.build.%v/path", GroupName)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ServiceBuildSpec   `json:"spec"`
	Status            ServiceBuildStatus `json:"status,omitempty"`
}

func (b *ServiceBuild) BuildIDLabel() (string, bool) {
	id, ok := b.Labels[BuildIDLabelKey]
	return id, ok
}

func (b *ServiceBuild) DefinitionURLLabel() (string, bool) {
	url, ok := b.Labels[ServiceBuildDefinitionURLLabelKey]
	return url, ok
}

func (b *ServiceBuild) DefinitionVersionLabel() (string, bool) {
	version, ok := b.Labels[ServiceBuildDefinitionVersionLabelKey]
	return version, ok
}

func (b *ServiceBuild) PathLabel() (tree.NodePath, error) {
	path, ok := b.Labels[ServiceBuildPathLabelKey]
	if !ok {
		return "", fmt.Errorf("service build did not contain service path label")
	}

	return tree.NodePathFromDomain(path)
}

func (b *ServiceBuild) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, b.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", b.Namespace))
	}

	path, err := b.PathLabel()
	if err != nil {
		path = tree.NodePath("unknown")
	}

	build := "unknown"
	if label, ok := b.BuildIDLabel(); ok {
		build = label
	}

	version := "unknown"
	if label, ok := b.DefinitionVersionLabel(); ok {
		version = label
	}

	definitionURL := "unknown definition URL"
	if label, ok := b.DefinitionURLLabel(); ok {
		definitionURL = label
	}

	return fmt.Sprintf(
		"service build %v (service %v in build %v, version %v of %v system %v)",
		b.Name,
		path.String(),
		build,
		version,
		definitionURL,
		systemID,
	)
}

// +k8s:deepcopy-gen=false
type ServiceBuildSpec struct {
	Components map[string]ServiceBuildSpecComponentBuildInfo `json:"components"`
}

// +k8s:deepcopy-gen=false
type ServiceBuildSpecComponentBuildInfo struct {
	DefinitionBlock block.ComponentBuild `json:"definitionBlock"`
}

type ServiceBuildStatus struct {
	// ServiceBuilds are immutable so no need for ObservedGeneration

	State   ServiceBuildState `json:"state"`
	Message string            `json:"message"`

	StartTimestamp      *metav1.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`

	// Maps a component name to the ComponentBuild.Name responsible for it
	ComponentBuilds map[string]string `json:"componentsBuilds"`

	// Maps a ComponentBuild.Name to the ComponentBuild.Status
	ComponentBuildStatuses map[string]ComponentBuildStatus `json:"componentBuildStatuses"`
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
