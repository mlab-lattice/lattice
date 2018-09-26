package system

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type BuildBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

func (b *BuildBackend) CreateFromPath(p tree.Path) (*v1.Build, error) {
	return b.create(&p, nil)
}

func (b *BuildBackend) CreateFromVersion(v v1.Version) (*v1.Build, error) {
	return b.create(nil, &v)
}

func (b *BuildBackend) create(p *tree.Path, v *v1.Version) (*v1.Build, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	build := b.backend.registry.CreateBuild(p, v, record)

	// run the build
	b.backend.controller.RunBuild(build, record)

	// copy the build so we don't return a pointer into the backend
	// so we can release the lock
	result := new(v1.Build)
	*result = *build

	return result, nil
}

func (b *BuildBackend) List() ([]v1.Build, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	var builds []v1.Build
	for _, build := range record.Builds {
		builds = append(builds, *build.Build)
	}

	return builds, nil
}

func (b *BuildBackend) Get(id v1.BuildID) (*v1.Build, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	build, ok := record.Builds[id]
	if !ok {
		return nil, v1.NewInvalidBuildIDError()
	}

	// copy the build so we don't return a pointer into the backend
	// so we can release the lock
	result := new(v1.Build)
	*result = *build.Build

	return result, nil
}

func (b *BuildBackend) Logs(
	id v1.BuildID,
	path tree.Path,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	_, err := b.Get(id)
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(strings.NewReader("this is a long line")), nil
}
