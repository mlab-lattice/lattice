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
	environment definitionv1.ContainerEnvironment,
) (*v1.Job, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	i, ok := record.definition.Get(path)
	if !ok {
		return nil, v1.NewInvalidPathError()
	}

	_, ok = i.Component.(*definitionv1.Job)
	if !ok {
		return nil, v1.NewInvalidPathError()
	}

	job := &v1.Job{
		ID:    v1.JobID(uuid.NewV4().String()),
		State: v1.JobStatePending,
		Path:  path,
	}

	record.jobs[job.ID] = job

	// run the job
	b.backend.controller.RunJob(job)

	// copy the build so we don't return a pointer into the backend
	// so we can release the lock
	result := new(v1.Job)
	*result = *job

	return result, nil
}

func (b *JobBackend) List() ([]v1.Job, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	var jobs []v1.Job
	for _, job := range record.jobs {
		jobs = append(jobs, *job)
	}

	return jobs, nil
}
func (b *JobBackend) Get(id v1.JobID) (*v1.Job, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	job, ok := record.jobs[id]
	if !ok {
		return nil, v1.NewInvalidJobIDError()
	}

	// copy the build so we don't return a pointer into the backend
	// so we can release the lock
	result := new(v1.Job)
	*result = *job

	return result, nil
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
