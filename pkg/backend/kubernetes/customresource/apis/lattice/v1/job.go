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
	ResourceSingularJob = "job"
	ResourcePluralJob   = "jobs"
	ResourceScopeJob    = apiextensionsv1beta1.NamespaceScoped
)

var (
	JobKind     = SchemeGroupVersion.WithKind("Job")
	JobListKind = SchemeGroupVersion.WithKind("JobList")

	// JobID label is the key that should be used in a label referencing a job's ID.
	JobIDLabelKey = fmt.Sprintf("job.%v/id", GroupName)

	// JobID label is the key that should be used for the path of the job.
	JobPathLabelKey = fmt.Sprintf("job.%v/path", GroupName)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Job struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              JobSpec `json:"spec"`
}

func (s *Job) Deleted() bool {
	return s.DeletionTimestamp != nil
}

func (s *Job) Description(namespacePrefix string) string {
	systemID, err := kubeutil.SystemID(namespacePrefix, s.Namespace)
	if err != nil {
		systemID = v1.SystemID(fmt.Sprintf("UNKNOWN (namespace: %v)", s.Namespace))
	}

	path, err := s.PathLabel()
	if err == nil {
		return fmt.Sprintf("job %v (%v in system %v)", s.Name, path, systemID)
	}

	return fmt.Sprintf("job %v (no path, system %v)", s.Name, systemID)
}

func (s *Job) PathLabel() (tree.NodePath, error) {
	path, ok := s.Labels[JobPathLabelKey]
	if !ok {
		return "", fmt.Errorf("job did not contain job path label")
	}

	return tree.NewNodePathFromDomain(path)
}

func (s *Job) NodePoolAnnotation() (NodePoolAnnotationValue, error) {
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
type JobSpec struct {
	Definition *definitionv1.Job `json:"definition"`

	// ContainerBuildArtifacts maps Sidecar names to the artifacts created by their build
	ContainerBuildArtifacts map[string]ContainerBuildArtifacts `json:"containerBuildArtifacts"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type JobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Job `json:"items"`
}
