package system

import (
	"io"
	"io/ioutil"
	"strings"
	"time"

	"fmt"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"github.com/satori/go.uuid"
)

type BuildBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

func (b *BuildBackend) Create(v v1.SystemVersion) (*v1.Build, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.Lock()
	defer record.recordLock.Unlock()

	// validate version
	if v != "1.0.0" {
		return nil, &v1.InvalidSystemVersionError{Version: string(v)}
	}

	build := newMockBuild(b.systemID, v)
	record.builds[build.ID] = build

	// run the build
	go runBuild(build)

	return build, nil
}

func (b *BuildBackend) List() ([]v1.Build, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	var builds []v1.Build
	for _, build := range record.builds {
		builds = append(builds, *build)
	}

	return builds, nil

}

func (b *BuildBackend) Get(buildID v1.BuildID) (*v1.Build, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	return b.backend.getBuildLocked(b.systemID, buildID)
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

func newMockBuild(systemID v1.SystemID, v v1.SystemVersion) *v1.Build {
	service1Path := tree.Path(fmt.Sprintf("/%s/api", systemID))
	build := &v1.Build{
		ID:      v1.BuildID(uuid.NewV4().String()),
		State:   v1.BuildStatePending,
		Version: v,
		Services: map[tree.Path]v1.ServiceBuild{
			service1Path: {
				ContainerBuild: v1.ContainerBuild{
					State: v1.ContainerBuildStatePending,
				},
			},
		},
	}

	return build
}

func runBuild(build *v1.Build) {
	// try to simulate reality by making things take a little longer. Sleep for a bit...
	time.Sleep(2 * time.Second)

	// change state to running
	build.State = v1.BuildStateRunning
	now := time.Now()
	build.StartTimestamp = &now

	// run service builds
	for sp, s := range build.Services {
		s.State = v1.ContainerBuildStateRunning
		s.StartTimestamp = &now
		s.ContainerBuild.State = v1.ContainerBuildStateRunning
		s.ContainerBuild.StartTimestamp = &now

		build.Services[sp] = s
	}

	// sleep
	fmt.Printf("Mock: Running build %s. Sleeping for 7 seconds\n", build.ID)
	time.Sleep(7 * time.Second)
	finishBuild(build)

}

func finishBuild(build *v1.Build) {
	// change state to succeeded
	now := time.Now()

	// finish service builds
	for sp, s := range build.Services {
		s.State = v1.ContainerBuildStateSucceeded
		s.CompletionTimestamp = &now
		s.ContainerBuild.CompletionTimestamp = &now
		s.ContainerBuild.State = v1.ContainerBuildStateSucceeded
		s.CompletionTimestamp = &now
		build.Services[sp] = s
	}

	build.CompletionTimestamp = &now
	build.State = v1.BuildStateSucceeded

	fmt.Printf("Build %s finished\n", build.ID)
}
