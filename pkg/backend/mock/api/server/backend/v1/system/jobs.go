package system

import (
	backendv1 "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/registry"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	"github.com/satori/go.uuid"
)

type JobBackend struct {
	backend *Backend
	system  v1.SystemID
}

func (b *JobBackend) Run(
	path tree.Path,
	command []string,
	environment definitionv1.ContainerExecEnvironment,
	numRetries *int32,
) (*v1.Job, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.system)
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

	record.Jobs[job.ID] = &registry.JobInfo{
		Job:  job,
		Runs: make(map[v1.JobRunID]v1.JobRun),
	}

	// run the job
	b.backend.controller.RunJob(job, record)

	// copy the build so we don't return a pointer into the backend
	// so we can release the lock
	return job.DeepCopy(), nil
}

func (b *JobBackend) List() ([]v1.Job, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.system)
	if err != nil {
		return nil, err
	}

	var jobs []v1.Job
	for _, job := range record.Jobs {
		jobs = append(jobs, *job.Job.DeepCopy())
	}

	return jobs, nil
}
func (b *JobBackend) Get(id v1.JobID) (*v1.Job, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.system)
	if err != nil {
		return nil, err
	}

	job, ok := record.Jobs[id]
	if !ok {
		return nil, v1.NewInvalidJobIDError()
	}

	// copy the build so we don't return a pointer into the backend
	// so we can release the lock
	return job.Job.DeepCopy(), nil
}

func (b *JobBackend) Runs(id v1.JobID) backendv1.SystemJobRunBackend {
	return &JobRunBackend{
		backend: b.backend,
		system:  b.system,
		job:     id,
	}
}
