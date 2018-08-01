package backend

import (
	"fmt"
	"io"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) RunJob(
	systemID v1.SystemID,
	path tree.Path,
	command []string,
	environment definitionv1.ContainerEnvironment,
) (*v1.Job, error) {
	// ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)

	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.JobPathLabelKey, selection.Equals, []string{path.ToDomain()})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for job %v/%v lookup: %v", namespace, path.String(), err)
	}
	selector = selector.Add(*requirement)

	jobs, err := kb.latticeClient.LatticeV1().Jobs(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	if len(jobs.Items) != 1 {
		err := fmt.Errorf("expected to find a job for %v in %v but found %v", path.String(), namespace, len(jobs.Items))
		return nil, err
	}

	job := jobs.Items[0]

	jobRun := &latticev1.JobRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewV4().String(),
			Labels: map[string]string{
				latticev1.JobRunPathLabelKey: path.ToDomain(),
			},
		},
		Spec: latticev1.JobRunSpec{
			Definition: job.Spec.Definition,

			Command:     command,
			Environment: environment,

			ContainerBuildArtifacts: job.Spec.ContainerBuildArtifacts,
		},
	}

	result, err := kb.latticeClient.LatticeV1().JobRuns(namespace).Create(jobRun)
	if err != nil {
		return nil, fmt.Errorf("error trying to create job run: %v", err)
	}

	externalJob, err := kb.transformJobRun(v1.JobID(result.Name), path, &result.Status, namespace)
	if err != nil {
		return nil, err
	}

	return &externalJob, nil
}

func (kb *KubernetesBackend) ListJobs(systemID v1.SystemID) ([]v1.Job, error) {
	// ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	jobRuns, err := kb.latticeClient.LatticeV1().JobRuns(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalJobs []v1.Job
	for _, jobRun := range jobRuns.Items {
		path, err := jobRun.PathLabel()
		if err != nil {
			return nil, err
		}

		externalJobRun, err := kb.transformJobRun(v1.JobID(jobRun.Name), path, &jobRun.Status, namespace)
		if err != nil {
			return nil, err
		}

		externalJobs = append(externalJobs, externalJobRun)
	}

	return externalJobs, nil
}

func (kb *KubernetesBackend) GetJob(systemID v1.SystemID, jobID v1.JobID) (*v1.Job, error) {
	// ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	jobRun, err := kb.latticeClient.LatticeV1().JobRuns(namespace).Get(string(jobID), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	path, err := jobRun.PathLabel()
	if err != nil {
		return nil, err
	}

	externalJob, err := kb.transformJobRun(v1.JobID(jobRun.Name), path, &jobRun.Status, namespace)
	if err != nil {
		return nil, err
	}

	return &externalJob, nil
}

func (kb *KubernetesBackend) JobLogs(
	systemID v1.SystemID,
	jobID v1.JobID,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	// Ensure the system exists
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)

	_, err := kb.latticeClient.LatticeV1().JobRuns(namespace).Get(string(jobID), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	pod, err := kb.findJobPod(jobID, namespace)

	if err != nil {
		return nil, err
	}

	container := kubeutil.UserMainContainerName
	if sidecar != nil {
		container = kubeutil.UserSidecarContainerName(*sidecar)
	}

	podLogOptions, err := toPodLogOptions(logOptions)
	if err != nil {
		return nil, err
	}
	podLogOptions.Container = container

	req := kb.kubeClient.CoreV1().Pods(namespace).GetLogs(pod.Name, podLogOptions)
	return req.Stream()

}

// findServicePod finds service pod by instance id or service's single pod if id was not specified
func (kb *KubernetesBackend) findJobPod(jobID v1.JobID, namespace string) (*corev1.Pod, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(latticev1.JobRunIDLabelKey, selection.Equals, []string{string(jobID)})
	if err != nil {
		return nil, fmt.Errorf("error creating requirement for %v/%v job lookup: %v", namespace, jobID, err)
	}

	selector = selector.Add(*requirement)
	pods, err := kb.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) > 1 {
		return nil, fmt.Errorf("found multiple pods for %v/%v", namespace, jobID)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods for %v/%v", namespace, jobID)
	}

	return &pods.Items[0], nil

}

func (kb *KubernetesBackend) transformJobRun(
	id v1.JobID,
	path tree.Path,
	status *latticev1.JobRunStatus,
	namespace string,
) (v1.Job, error) {
	state, err := getJobRunStateState(status.State)
	if err != nil {
		return v1.Job{}, err
	}

	var startTimestamp *time.Time
	if status.StartTimestamp != nil {
		startTimestamp = &status.StartTimestamp.Time
	}

	var completionTimestamp *time.Time
	if status.CompletionTimestamp != nil {
		completionTimestamp = &status.CompletionTimestamp.Time
	}

	job := v1.Job{
		ID:    id,
		Path:  path,
		State: state,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,
	}
	return job, nil
}

func getJobRunStateState(state latticev1.JobRunState) (v1.JobState, error) {
	switch state {
	case latticev1.JobRunStatePending:
		return v1.JobStatePending, nil
	case latticev1.JobRunStateDeleting:
		return v1.JobStateDeleting, nil

	case latticev1.JobRunStateQueued:
		return v1.JobStateQueued, nil
	case latticev1.JobRunStateRunning:
		return v1.JobStateRunning, nil
	case latticev1.JobRunStateSucceeded:
		return v1.JobStateSucceeded, nil
	case latticev1.JobRunStateFailed:
		return v1.JobStateFailed, nil
	default:
		return "", fmt.Errorf("invalid job state: %v", state)
	}
}
