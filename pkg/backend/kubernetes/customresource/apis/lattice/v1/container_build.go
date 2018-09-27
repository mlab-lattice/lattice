package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ContainerBuildKind     = SchemeGroupVersion.WithKind("ContainerBuild")
	ContainerBuildListKind = SchemeGroupVersion.WithKind("ContainerBuildList")

	ContainerBuildIDLabelKey             = fmt.Sprintf("containerbuild.%v/id", GroupName)
	ContainerBuildDefinitionHashLabelKey = fmt.Sprintf("containerbuild.%v/definition-hash", GroupName)

	ContainerBuildJobDockerImageFQNAnnotationKey = fmt.Sprintf("containerbuild.%v/docker-image-fqn", GroupName)

	ContainerBuildFailureInfoAnnotationKey       = fmt.Sprintf("containerbuild.%v/failure-info", GroupName)
	ContainerBuildLastObservedPhaseAnnotationKey = fmt.Sprintf("containerbuild.%v/last-observed-phase", GroupName)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ContainerBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ContainerBuildSpec   `json:"spec"`
	Status            ContainerBuildStatus `json:"status"`
}

func (b *ContainerBuild) DefinitionHashLabel() (string, bool) {
	hash, ok := b.Labels[ContainerBuildDefinitionHashLabelKey]
	return hash, ok
}

func (b *ContainerBuild) FailureInfoAnnotation() (*v1.ContainerBuildFailureInfo, error) {
	infoStr, ok := b.Annotations[ContainerBuildFailureInfoAnnotationKey]
	if !ok {
		return nil, nil
	}

	failureInfo := v1.ContainerBuildFailureInfo{}
	err := json.Unmarshal([]byte(infoStr), &failureInfo)
	if err != nil {
		return nil, err
	}

	return &failureInfo, nil
}

func (b *ContainerBuild) LastObservedPhaseAnnotation() (v1.ContainerBuildPhase, bool) {
	phase, ok := b.Annotations[ContainerBuildLastObservedPhaseAnnotationKey]
	return v1.ContainerBuildPhase(phase), ok
}

func (b *ContainerBuild) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, b.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", b.Namespace))
	}

	return fmt.Sprintf("component build %v (system %v)", b.Name, systemID)
}

// +k8s:deepcopy-gen=false
type ContainerBuildSpec struct {
	Definition *definitionv1.ContainerBuild `json:"definition"`
}

type ContainerBuildStatus struct {
	// ContainerBuilds are immutable so no need for ObservedGeneration

	State       ComponentBuildState           `json:"state"`
	FailureInfo *v1.ContainerBuildFailureInfo `json:"failureInfo,omitempty"`

	StartTimestamp      *metav1.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`

	Artifacts         *ContainerBuildArtifacts `json:"artifacts,omitempty"`
	LastObservedPhase *v1.ContainerBuildPhase  `json:"lastObservedPhase,omitempty"`
}

type ComponentBuildState string

const (
	ContainerBuildStatePending   ComponentBuildState = ""
	ContainerBuildStateQueued    ComponentBuildState = "queued"
	ContainerBuildStateRunning   ComponentBuildState = "running"
	ContainerBuildStateSucceeded ComponentBuildState = "succeeded"
	ContainerBuildStateFailed    ComponentBuildState = "failed"
)

type ContainerBuildArtifacts struct {
	DockerImageFQN string `json:"dockerImageFqn"`
}

type WorkloadContainerBuildArtifacts struct {
	MainContainer ContainerBuildArtifacts            `json:"mainContainer"`
	Sidecars      map[string]ContainerBuildArtifacts `json:"sidecars"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ContainerBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ContainerBuild `json:"items"`
}
