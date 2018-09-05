package system

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"io"
	"io/ioutil"
	"strings"

	"github.com/satori/go.uuid"
)

type BuildBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

func (b *BuildBackend) Create(v v1.SystemVersion) (*v1.Build, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	// validate version
	if v != "1.0.0" {
		return nil, v1.NewInvalidSystemVersionError()
	}

	build := newBuild(v)
	record.builds[build.ID] = build

	// run the build
	b.backend.controller.RunBuild(build)

	// copy the build so we don't return a pointer into the backend
	// so we can release the lock
	result := new(v1.Build)
	*result = *build

	return result, nil
}

func (b *BuildBackend) List() ([]v1.Build, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	var builds []v1.Build
	for _, build := range record.builds {
		builds = append(builds, *build)
	}

	return builds, nil
}

func (b *BuildBackend) Get(id v1.BuildID) (*v1.Build, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	build, ok := record.builds[id]
	if !ok {
		return nil, v1.NewInvalidBuildIDError()
	}

	// copy the build so we don't return a pointer into the backend
	// so we can release the lock
	result := new(v1.Build)
	*result = *build

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

func newBuild(v v1.SystemVersion) *v1.Build {
	service1Path := tree.Path("/api")
	build := &v1.Build{
		ID:      v1.BuildID(uuid.NewV4().String()),
		State:   v1.BuildStatePending,
		Version: v,
		Services: map[tree.Path]v1.ServiceBuild{
			service1Path: {
				ContainerBuild: v1.ContainerBuild{
					ID:    v1.ContainerBuildID(uuid.NewV4().String()),
					State: v1.ContainerBuildStatePending,
				},
			},
		},
	}

	return build
}
