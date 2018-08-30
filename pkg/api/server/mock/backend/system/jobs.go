package system

import (
	"fmt"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
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
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.Lock()
	defer record.recordLock.Unlock()

	job := &v1.Job{
		ID:    v1.JobID(uuid.NewV4().String()),
		State: v1.JobStatePending,
		Path:  path,
	}

	record.jobs[job.ID] = job

	// run the job
	go runJob(job)

	return job, nil
}

func (b *JobBackend) List() ([]v1.Job, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	var jobs []v1.Job
	for _, job := range record.jobs {
		jobs = append(jobs, *job)
	}

	return jobs, nil
}
func (b *JobBackend) Get(id v1.JobID) (*v1.Job, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	job, ok := record.jobs[id]
	if !ok {
		return nil, v1.NewInvalidJobIDError(id)
	}

	return job, nil
}

func (b *Backend) Logs(
	id v1.JobID,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader("this is a long line")), nil
}

func runJob(job *v1.Job) {
	// try to simulate reality by making things take a little longer. Sleep for a bit...
	time.Sleep(2 * time.Second)

	// change state to running
	job.State = v1.JobStateRunning
	now := time.Now()
	job.StartTimestamp = &now

	// sleep
	fmt.Printf("running job %s. Sleeping for 7 seconds\n", job.ID)
	time.Sleep(7 * time.Second)
	finishJob(job)
}

func finishJob(job *v1.Job) {
	// change state to succeeded
	now := time.Now()

	job.CompletionTimestamp = &now
	job.State = v1.JobStateSucceeded

	fmt.Printf("Job %s finished\n", job.ID)
}
