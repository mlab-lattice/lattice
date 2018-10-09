package system

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type JobRunBackend struct {
	backend *Backend
	system  v1.SystemID
	job     v1.JobID
}

func (b *JobRunBackend) List() ([]v1.JobRun, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.system)
	if err != nil {
		return nil, err
	}

	info, ok := record.Jobs[b.job]
	if !ok {
		return nil, v1.NewInvalidJobIDError()
	}

	runs := make([]v1.JobRun, len(info.Runs))
	i := 0
	for _, run := range info.Runs {
		runs[i] = *run.DeepCopy()
		i++
	}

	return runs, nil
}

func (b *JobRunBackend) Get(id v1.JobRunID) (*v1.JobRun, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.system)
	if err != nil {
		return nil, err
	}

	info, ok := record.Jobs[b.job]
	if !ok {
		return nil, v1.NewInvalidJobIDError()
	}

	run, ok := info.Runs[id]
	if !ok {
		return nil, v1.NewInvalidJobRunIDError()
	}

	return run.DeepCopy(), nil
}

func (b *JobRunBackend) Logs(
	id v1.JobRunID,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	_, err := b.Get(id)
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(strings.NewReader("this is a long line")), nil
}
