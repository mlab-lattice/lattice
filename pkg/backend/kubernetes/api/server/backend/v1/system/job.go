package system

import (
	"fmt"

	backendv1 "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mlab-lattice/lattice/pkg/util/time"
	"github.com/satori/go.uuid"
)

type jobBackend struct {
	backend *Backend
	system  v1.SystemID
}

func (b *jobBackend) Run(
	path tree.Path,
	command []string,
	environment definitionv1.ContainerExecEnvironment,
	numRetries *int32,
) (*v1.Job, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)
	definition, artifacts, err := b.getJobInformation(path, namespace)
	if err != nil {
		return nil, err
	}

	job := &latticev1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewV4().String(),
			Labels: map[string]string{
				latticev1.JobPathLabelKey: path.ToDomain(),
			},
		},
		Spec: latticev1.JobSpec{
			Definition: *definition,

			Command:     command,
			Environment: environment,

			NumRetries: numRetries,

			ContainerBuildArtifacts: artifacts,
		},
	}

	result, err := b.backend.latticeClient.LatticeV1().Jobs(namespace).Create(job)
	if err != nil {
		return nil, fmt.Errorf("error trying to create job run: %v", err)
	}

	externalJob, err := b.transformJob(v1.JobID(result.Name), path, &result.Status, namespace)
	if err != nil {
		return nil, err
	}

	return &externalJob, nil
}

func (b *jobBackend) List() ([]v1.Job, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)
	jobs, err := b.backend.latticeClient.LatticeV1().Jobs(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalJobs []v1.Job
	for _, job := range jobs.Items {
		path, err := job.PathLabel()
		if err != nil {
			return nil, err
		}

		externalJob, err := b.transformJob(v1.JobID(job.Name), path, &job.Status, namespace)
		if err != nil {
			return nil, err
		}

		externalJobs = append(externalJobs, externalJob)
	}

	return externalJobs, nil
}

func (b *jobBackend) Get(id v1.JobID) (*v1.Job, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)
	jobRun, err := b.backend.latticeClient.LatticeV1().Jobs(namespace).Get(string(id), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	path, err := jobRun.PathLabel()
	if err != nil {
		return nil, err
	}

	externalJob, err := b.transformJob(v1.JobID(jobRun.Name), path, &jobRun.Status, namespace)
	if err != nil {
		return nil, err
	}

	return &externalJob, nil
}

func (b *jobBackend) Instances(id v1.JobID) backendv1.SystemJobRunBackend {
	return &jobRunBackend{
		backend: b.backend,
		system:  b.system,
		job:     id,
	}
}

func (b *jobBackend) getJobInformation(
	path tree.Path,
	namespace string,
) (
	*definitionv1.Job,
	latticev1.WorkloadContainerBuildArtifacts,
	error,
) {
	system, err := b.backend.getLatticeV1System(b.system)
	if err != nil {
		return nil, latticev1.WorkloadContainerBuildArtifacts{}, err
	}

	if system.Spec.Definition == nil {
		return nil, latticev1.WorkloadContainerBuildArtifacts{}, v1.NewInvalidPathError()
	}

	info, ok := system.Spec.Definition.V1().Get(path)
	if !ok {
		return nil, latticev1.WorkloadContainerBuildArtifacts{}, v1.NewInvalidPathError()
	}

	job, ok := info.Component.(*definitionv1.Job)
	if !ok {
		return nil, latticev1.WorkloadContainerBuildArtifacts{}, v1.NewInvalidComponentTypeError()
	}

	if system.Spec.WorkloadBuildArtifacts == nil {
		err := fmt.Errorf("%v has non-nil definition but nil workload build artifacts", system.Description())
		return nil, latticev1.WorkloadContainerBuildArtifacts{}, err
	}

	artifacts, ok := system.Spec.WorkloadBuildArtifacts.Get(path)
	if !ok {
		err := fmt.Errorf("%v has job %v in definition but no workload build artifacts", system.Description(), path.String())
		return nil, latticev1.WorkloadContainerBuildArtifacts{}, err
	}

	return job, artifacts, nil
}

func (b *jobBackend) transformJob(
	id v1.JobID,
	path tree.Path,
	status *latticev1.JobStatus,
	namespace string,
) (v1.Job, error) {
	state, err := getJobState(status.State)
	if err != nil {
		return v1.Job{}, err
	}

	var startTimestamp *time.Time
	if status.StartTimestamp != nil {
		startTimestamp = time.New(status.StartTimestamp.Time)
	}

	var completionTimestamp *time.Time
	if status.CompletionTimestamp != nil {
		completionTimestamp = time.New(status.CompletionTimestamp.Time)
	}

	job := v1.Job{
		ID: id,

		Path: path,

		Status: v1.JobStatus{
			State: state,

			StartTimestamp:      startTimestamp,
			CompletionTimestamp: completionTimestamp,
		},
	}
	return job, nil
}

func getJobState(state latticev1.JobState) (v1.JobState, error) {
	switch state {
	case latticev1.JobStatePending:
		return v1.JobStatePending, nil
	case latticev1.JobStateDeleting:
		return v1.JobStateDeleting, nil

	case latticev1.JobStateQueued:
		return v1.JobStateQueued, nil
	case latticev1.JobStateRunning:
		return v1.JobStateRunning, nil
	case latticev1.JobStateSucceeded:
		return v1.JobStateSucceeded, nil
	case latticev1.JobStateFailed:
		return v1.JobStateFailed, nil
	default:
		return "", fmt.Errorf("invalid job state: %v", state)
	}
}
