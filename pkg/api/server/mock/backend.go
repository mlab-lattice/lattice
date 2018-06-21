package mock

import (
	"io"

	"time"

	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/satori/go.uuid"
)

type MockBackend struct {
	systemRegistry map[v1.SystemID]*SystemRecord
}

func newMockBackend() *MockBackend {
	return &MockBackend{
		systemRegistry: make(map[v1.SystemID]*SystemRecord),
	}
}

type SystemRecord struct {
	system  *v1.System
	builds  map[v1.BuildID]*v1.Build
	deploys map[v1.DeployID]*v1.Deploy
}

// newSystemRecord
func newSystemRecord(system *v1.System) *SystemRecord {
	return &SystemRecord{
		system:  system,
		builds:  make(map[v1.BuildID]*v1.Build),
		deploys: make(map[v1.DeployID]*v1.Deploy),
	}
}

// Systems
func (backend *MockBackend) CreateSystem(systemID v1.SystemID, definitionURL string) (*v1.System, error) {

	if _, exists := backend.systemRegistry[systemID]; exists {
		return nil, v1.NewSystemAlreadyExistsError(systemID)
	}

	// create system
	system := &v1.System{
		ID:            systemID,
		State:         v1.SystemStateStable,
		DefinitionURL: definitionURL,
	}
	// register it with in memory map
	backend.systemRegistry[systemID] = newSystemRecord(system)
	return system, nil
}

func (backend *MockBackend) ListSystems() ([]v1.System, error) {
	systems := []v1.System{}
	for _, v := range backend.systemRegistry {
		systems = append(systems, *v.system)
	}
	return systems, nil
}

func (backend *MockBackend) GetSystem(systemID v1.SystemID) (*v1.System, error) {
	systemRecord, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	return systemRecord.system, nil
}

func (backend *MockBackend) DeleteSystem(systemID v1.SystemID) error {
	if _, exists := backend.systemRegistry[systemID]; !exists {
		return v1.NewInvalidSystemIDError(systemID)
	}
	delete(backend.systemRegistry, systemID)
	return nil
}

// Builds
func (backend *MockBackend) Build(systemID v1.SystemID, definitionRoot tree.Node, v v1.SystemVersion) (*v1.Build, error) {

	// ensure the system exists
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	build := newMockBuild(systemID, v)

	record.builds[build.ID] = build
	// run the build
	go runBuild(build)
	return build, nil
}

func (backend *MockBackend) ListBuilds(systemID v1.SystemID) ([]v1.Build, error) {
	// ensure the system exists
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	builds := []v1.Build{}
	for _, build := range record.builds {
		builds = append(builds, *build)
	}
	return builds, nil

}

func (backend *MockBackend) GetBuild(systemID v1.SystemID, buildID v1.BuildID) (*v1.Build, error) {
	// ensure the system exists
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	build, exists := record.builds[buildID]

	if !exists {
		return nil, v1.NewInvalidBuildIDError(buildID)
	}

	return build, nil
}

func (backend *MockBackend) BuildLogs(systemID v1.SystemID, buildID v1.BuildID, path tree.NodePath, component string,
	logOptions *v1.ContainerLogOptions) (io.ReadCloser, error) {
	return nil, nil
}

// Deploys
func (backend *MockBackend) DeployBuild(systemID v1.SystemID, buildID v1.BuildID) (*v1.Deploy, error) {
	_, err := backend.GetBuild(systemID, buildID)
	if err != nil {
		return nil, err
	}

	deploy := &v1.Deploy{
		ID:      v1.DeployID(uuid.NewV4().String()),
		BuildID: buildID,
		State:   v1.DeployStatePending,
	}

	record := backend.systemRegistry[systemID]
	record.deploys[deploy.ID] = deploy

	return deploy, nil
}

func (backend *MockBackend) DeployVersion(systemID v1.SystemID, definitionRoot tree.Node, version v1.SystemVersion) (*v1.Deploy, error) {
	// this ensures the system is created as well
	build, err := backend.Build(systemID, definitionRoot, version)
	if err != nil {
		return nil, err
	}

	return backend.DeployBuild(systemID, build.ID)
}

func (backend *MockBackend) ListDeploys(systemID v1.SystemID) ([]v1.Deploy, error) {
	// ensure the system exists
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	deploys := []v1.Deploy{}
	for _, deploy := range record.deploys {
		deploys = append(deploys, *deploy)
	}
	return deploys, nil
}

func (backend *MockBackend) GetDeploy(systemID v1.SystemID, deployID v1.DeployID) (*v1.Deploy, error) {
	// ensure the system exists
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	deploy, exists := record.deploys[deployID]

	if !exists {
		return nil, v1.NewInvalidDeployIDError(deployID)
	}

	return deploy, nil
}

// Teardown
func (backend *MockBackend) TearDown(v1.SystemID) (*v1.Teardown, error) {
	return nil, nil
}

func (backend *MockBackend) ListTeardowns(v1.SystemID) ([]v1.Teardown, error) {
	return nil, nil
}

func (backend *MockBackend) GetTeardown(v1.SystemID, v1.TeardownID) (*v1.Teardown, error) {
	return nil, nil
}

// Services
func (backend *MockBackend) ListServices(v1.SystemID) ([]v1.Service, error) {
	return nil, nil
}

func (backend *MockBackend) GetService(v1.SystemID, v1.ServiceID) (*v1.Service, error) {
	return nil, nil
}

func (backend *MockBackend) GetServiceByPath(v1.SystemID, tree.NodePath) (*v1.Service, error) {
	return nil, nil
}

func (backend *MockBackend) ServiceLogs(systemID v1.SystemID, serviceID v1.ServiceID, component string,
	instance string, logOptions *v1.ContainerLogOptions) (io.ReadCloser, error) {
	return nil, nil
}

// Secrets
func (backend *MockBackend) ListSystemSecrets(v1.SystemID) ([]v1.Secret, error) {
	return nil, nil
}

func (backend *MockBackend) GetSystemSecret(systemID v1.SystemID, path tree.NodePath, name string) (*v1.Secret, error) {
	return nil, nil
}

func (backend *MockBackend) SetSystemSecret(systemID v1.SystemID, path tree.NodePath, name, value string) error {
	return nil
}

func (backend *MockBackend) UnsetSystemSecret(systemID v1.SystemID, path tree.NodePath, name string) error {
	return nil
}

func (backend *MockBackend) ListNodePools(v1.SystemID) ([]v1.NodePool, error) {
	return nil, nil
}

func (backend *MockBackend) GetNodePool(v1.SystemID, v1.NodePoolPath) (*v1.NodePool, error) {
	return nil, nil
}

// helpers

func newMockBuild(systemID v1.SystemID, v v1.SystemVersion) *v1.Build {
	service1Path := tree.NodePath(fmt.Sprintf("/%s/api", systemID))
	build := &v1.Build{
		ID:      v1.BuildID(uuid.NewV4().String()),
		State:   v1.BuildStatePending,
		Version: v,
		Services: map[tree.NodePath]v1.ServiceBuild{
			service1Path: {
				State: v1.ServiceBuildStatePending,
				Components: map[string]v1.ComponentBuild{
					"api": {
						State: v1.ComponentBuildStatePending,
					},
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
	startTime := time.Now()
	build.StartTimestamp = &startTime

	// run service builds
	for sp, s := range build.Services {
		s.State = v1.ServiceBuildStateRunning
		s.StartTimestamp = &startTime
		for cp, c := range s.Components {
			c.State = v1.ComponentBuildStateRunning
			c.StartTimestamp = &startTime
			s.Components[cp] = c
		}

		build.Services[sp] = s
	}

	// sleep
	fmt.Printf("Mock: Running build %s. Sleeping for 20 seconds\n", build.ID)
	time.Sleep(20 * time.Second)

	// wake up
	// change state to succeeded
	endTime := time.Now()

	// finish service builds
	for sp, s := range build.Services {
		s.State = v1.ServiceBuildStateSucceeded
		s.CompletionTimestamp = &endTime
		for cp, c := range s.Components {
			c.CompletionTimestamp = &startTime
			c.State = v1.ComponentBuildStateSucceeded
			s.Components[cp] = c
		}
		s.CompletionTimestamp = &startTime
		build.Services[sp] = s
	}

	build.CompletionTimestamp = &endTime
	build.State = v1.BuildStateSucceeded
	fmt.Printf("Build %s finished\n", build.ID)
}
