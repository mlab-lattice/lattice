package system

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"strings"
)

type JobBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

func (b *JobBackend) Run(
	path tree.Path,
	command []string,
	environment definitionv1.ContainerExecEnvironment,
	numRetries *int32,
) (*v1.Job, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	i, ok := record.Definition.Get(path)
	if !ok {
		return nil, v1.NewInvalidPathError()
	}

	_, ok = i.Component.(*definitionv1.Job)
	if !ok {
		return nil, v1.NewInvalidPathError()
	}

	job := &v1.Job{
		ID:   v1.JobID(uuid.NewV4().String()),
		Path: path,
		Status: v1.JobStatus{
			State: v1.JobStatePending,
		},
	}

	record.Jobs[job.ID] = job

	// run the job
	b.backend.controller.RunJob(job)

	// copy the build so we don't return a pointer into the backend
	// so we can release the lock
	return job.DeepCopy(), nil
}

func (b *JobBackend) List() ([]v1.Job, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	var jobs []v1.Job
	for _, job := range record.Jobs {
		jobs = append(jobs, *job.DeepCopy())
	}

	return jobs, nil
}
func (b *JobBackend) Get(id v1.JobID) (*v1.Job, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	job, ok := record.Jobs[id]
	if !ok {
		return nil, v1.NewInvalidJobIDError()
	}

	// copy the build so we don't return a pointer into the backend
	// so we can release the lock
	return job.DeepCopy(), nil
}

func (b *JobBackend) Logs(
	id v1.JobID,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	_, err := b.Get(id)
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(strings.NewReader("this is a long line")), nil
}
