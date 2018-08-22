package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularJobRun = "jobrun"
	ResourcePluralJobRun   = "jobruns"
	ResourceScopeJobRun    = apiextensionsv1beta1.NamespaceScoped
)

var (
	JobRunKind     = SchemeGroupVersion.WithKind("JobRun")
	JobRunListKind = SchemeGroupVersion.WithKind("JobRunList")

	// JobRunID label is the key that should be used in a label referencing a jobRun's ID.
	JobRunIDLabelKey = fmt.Sprintf("jobrun.%v/id", GroupName)

	// JobRunID label is the key that should be used for the path of the jobRun.
	JobRunPathLabelKey = fmt.Sprintf("jobrun.%v/path", GroupName)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type JobRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              JobRunSpec   `json:"spec"`
	Status            JobRunStatus `json:"status,omitempty"`
}

func (s *JobRun) Deleted() bool {
	return s.DeletionTimestamp != nil
}

func (s *JobRun) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, s.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", s.Namespace))
	}

	path, err := s.PathLabel()
	if err == nil {
		return fmt.Sprintf("job run %v (%v in system %v)", s.Name, path, systemID)
	}

	return fmt.Sprintf("job run %v (no path, system %v)", s.Name, systemID)
}

func (s *JobRun) PathLabel() (tree.Path, error) {
	path, ok := s.Labels[JobRunPathLabelKey]
	if !ok {
		return "", fmt.Errorf("job run did not contain job run path label")
	}

	return tree.NewPathFromDomain(path)
}

func (s *JobRun) NodePoolAnnotation() (NodePoolAnnotationValue, error) {
	annotation := make(NodePoolAnnotationValue)
	existingAnnotationString, ok := s.Annotations[NodePoolWorkloadAnnotationKey]
	if ok {
		err := json.Unmarshal([]byte(existingAnnotationString), &annotation)
		if err != nil {
			return nil, err
		}
	}

	return annotation, nil
}

// +k8s:deepcopy-gen=false
type JobRunSpec struct {
	Definition *definitionv1.Job `json:"definition"`

	NumRetries  *int32                            `json:"numRetries"`
	Command     []string                          `json:"command"`
	Environment definitionv1.ContainerEnvironment `json:"environment"`

	// ContainerBuildArtifacts maps container names to the artifacts created by their build
	ContainerBuildArtifacts map[string]ContainerBuildArtifacts `json:"containerBuildArtifacts"`
}

type JobRunStatus struct {
	State JobRunState `json:"state"`

	StartTimestamp      *metav1.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *metav1.Time `json:"completionTimestamp,omitempty"`
}

type JobRunState string

const (
	JobRunStatePending  JobRunState = ""
	JobRunStateDeleting JobRunState = "deleting"

	JobRunStateQueued    JobRunState = "queued"
	JobRunStateRunning   JobRunState = "running"
	JobRunStateSucceeded JobRunState = "succeeded"
	JobRunStateFailed    JobRunState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type JobRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []JobRun `json:"items"`
}
