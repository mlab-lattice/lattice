package mock

import (
	"io"

	"time"

	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/satori/go.uuid"

	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
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
	system    *v1.System
	builds    map[v1.BuildID]*v1.Build
	deploys   map[v1.DeployID]*v1.Deploy
	teardowns map[v1.TeardownID]*v1.Teardown
	secrets   []v1.Secret
	nodePools []v1.NodePool
}

// newSystemRecord
func newSystemRecord(system *v1.System) *SystemRecord {
	return &SystemRecord{
		system:    system,
		builds:    make(map[v1.BuildID]*v1.Build),
		deploys:   make(map[v1.DeployID]*v1.Deploy),
		teardowns: make(map[v1.TeardownID]*v1.Teardown),
		secrets:   []v1.Secret{},
		nodePools: []v1.NodePool{},
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

func (backend *MockBackend) Build(
	systemID v1.SystemID,
	def *definitionv1.SystemNode,
	v v1.SystemVersion) (*v1.Build, error) {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	build := backend.newMockBuild(systemID, v)

	record.builds[build.ID] = build

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

func (backend *MockBackend) BuildLogs(
	systemID v1.SystemID,
	buildID v1.BuildID,
	path tree.NodePath,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	return nil, nil
}

// Deploys
func (backend *MockBackend) DeployBuild(systemID v1.SystemID, buildID v1.BuildID) (*v1.Deploy, error) {
	build, err := backend.GetBuild(systemID, buildID)
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

	// run the build
	go backend.runBuild(build)

	return deploy, nil
}

func (backend *MockBackend) DeployVersion(
	systemID v1.SystemID,
	def *definitionv1.SystemNode,
	version v1.SystemVersion) (*v1.Deploy, error) {
	// this ensures the system is created as well
	build, err := backend.Build(systemID, def, version)
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
func (backend *MockBackend) TearDown(systemID v1.SystemID) (*v1.Teardown, error) {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	teardown := &v1.Teardown{
		ID:    v1.TeardownID(uuid.NewV4().String()),
		State: v1.TeardownStatePending,
	}

	record.teardowns[teardown.ID] = teardown
	// run the teardown
	go backend.runTeardown(teardown)
	return teardown, nil
}

func (backend *MockBackend) ListTeardowns(systemID v1.SystemID) ([]v1.Teardown, error) {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	teardowns := []v1.Teardown{}
	for _, teardown := range record.teardowns {
		teardowns = append(teardowns, *teardown)
	}
	return teardowns, nil
}

func (backend *MockBackend) GetTeardown(systemID v1.SystemID, teardownID v1.TeardownID) (*v1.Teardown, error) {
	// ensure the system exists
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	teardown, exists := record.teardowns[teardownID]

	if !exists {
		return nil, v1.NewInvalidTeardownIDError(teardownID)
	}

	return teardown, nil
}

// Services
func (backend *MockBackend) ListServices(systemID v1.SystemID) ([]v1.Service, error) {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	services := []v1.Service{}
	for _, service := range record.system.Services {
		services = append(services, service)
	}
	return services, nil
}

func (backend *MockBackend) GetService(systemID v1.SystemID, serviceID v1.ServiceID) (*v1.Service, error) {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	for _, service := range record.system.Services {
		if service.ID == serviceID {
			return &service, nil
		}
	}

	return nil, v1.NewInvalidServiceIDError(serviceID)
}

func (backend *MockBackend) GetServiceByPath(systemID v1.SystemID, path tree.NodePath) (*v1.Service, error) {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	service, exists := record.system.Services[path]

	if !exists {
		return nil, v1.NewInvalidServicePathError(path)
	}

	return &service, nil
}

func (backend *MockBackend) ServiceLogs(
	systemID v1.SystemID,
	serviceID v1.ServiceID,
	sidecar *string,
	instance string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	return nil, nil
}

// Secrets
func (backend *MockBackend) ListSystemSecrets(systemID v1.SystemID) ([]v1.Secret, error) {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	return record.secrets, nil
}

func (backend *MockBackend) GetSystemSecret(systemID v1.SystemID, path tree.NodePath, name string) (*v1.Secret, error) {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	for _, secret := range record.secrets {
		if secret.Path == path && secret.Name == name {
			scrt := &secret
			return scrt, nil
		}
	}

	return nil, v1.NewInvalidSystemSecretError(path, name)
}

func (backend *MockBackend) SetSystemSecret(systemID v1.SystemID, path tree.NodePath, name, value string) error {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return v1.NewInvalidSystemIDError(systemID)
	}

	secret := v1.Secret{
		Path:  path,
		Name:  name,
		Value: value,
	}

	record.secrets = append(record.secrets, secret)

	return nil
}

func (backend *MockBackend) UnsetSystemSecret(systemID v1.SystemID, path tree.NodePath, name string) error {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return v1.NewInvalidSystemIDError(systemID)
	}

	for i, secret := range record.secrets {
		if secret.Path == path && secret.Name == name {
			// delete secret
			record.secrets = append(record.secrets[:i], record.secrets[i+1:]...)

			return nil
		}
	}

	return v1.NewInvalidSystemSecretError(path, name)
}

func (backend *MockBackend) ListNodePools(systemID v1.SystemID) ([]v1.NodePool, error) {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	return record.nodePools, nil
}

func (backend *MockBackend) GetNodePool(systemID v1.SystemID, path v1.NodePoolPath) (*v1.NodePool, error) {
	record, exists := backend.systemRegistry[systemID]
	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	for _, nodePool := range record.nodePools {
		if nodePool.Path == path.String() {
			np := &nodePool
			return np, nil
		}
	}
	return nil, nil
}

// Jobs
func (backend *MockBackend) RunJob(systemID v1.SystemID, path tree.NodePath, command []string,
	environment definitionv1.ContainerEnvironment,
) (*v1.Job, error) {
	return nil, nil
}

func (backend *MockBackend) ListJobs(v1.SystemID) ([]v1.Job, error) {
	return nil, nil
}
func (backend *MockBackend) GetJob(v1.SystemID, v1.JobID) (*v1.Job, error) {
	return nil, nil
}
func (backend *MockBackend) JobLogs(systemID v1.SystemID, jobID v1.JobID, sidecar *string, logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	return nil, nil
}

// helpers

func (backend *MockBackend) getSystemRecordForBuild(buildID v1.BuildID) *SystemRecord {
	for _, systemRecord := range backend.systemRegistry {
		if _, exists := systemRecord.builds[buildID]; exists {
			return systemRecord
		}
	}

	return nil
}

func (backend *MockBackend) getDeployForBuild(buildID v1.BuildID) *v1.Deploy {
	for _, systemRecord := range backend.systemRegistry {
		for _, deploy := range systemRecord.deploys {
			if deploy.BuildID == buildID {
				return deploy
			}
		}
	}
	return nil
}

func (backend *MockBackend) newMockBuild(systemID v1.SystemID, v v1.SystemVersion) *v1.Build {
	service1Path := tree.NodePath(fmt.Sprintf("/%s/api", systemID))
	build := &v1.Build{
		ID:      v1.BuildID(uuid.NewV4().String()),
		State:   v1.BuildStatePending,
		Version: v,
		Services: map[tree.NodePath]v1.ServiceBuild{
			service1Path: {
				ContainerBuild: v1.ContainerBuild{
					State: v1.ContainerBuildStatePending,
				},
			},
		},
	}

	return build
}

func (backend *MockBackend) runBuild(build *v1.Build) {
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

	// run associated deploy
	deploy := backend.getDeployForBuild(build.ID)
	deploy.State = v1.DeployStateInProgress

	// sleep
	fmt.Printf("Mock: Running build %s. Sleeping for 7 seconds\n", build.ID)
	time.Sleep(7 * time.Second)
	backend.finishBuild(build)

}

func (backend *MockBackend) finishBuild(build *v1.Build) {
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

	// succeed associated deploy
	deploy := backend.getDeployForBuild(build.ID)
	deploy.State = v1.DeployStateSucceeded

	fmt.Printf("Build %s finished\n", build.ID)

	// update system services...
	systemRecord := backend.getSystemRecordForBuild(build.ID)
	services := make(map[tree.NodePath]v1.Service)
	for _, build := range systemRecord.builds {
		for path, _ := range build.Services {
			services[path] = v1.Service{
				ID:                 v1.ServiceID(uuid.NewV4().String()),
				State:              v1.ServiceStateStable,
				Instances:          []string{uuid.NewV4().String()},
				Path:               path,
				AvailableInstances: 1,
			}
		}
		break
	}

	systemRecord.system.Services = services

}

func (backend *MockBackend) getSystemRecordForTeardown(teardownID v1.TeardownID) *SystemRecord {
	for _, systemRecord := range backend.systemRegistry {
		if _, exists := systemRecord.teardowns[teardownID]; exists {
			return systemRecord
		}
	}
	return nil
}

func (backend *MockBackend) runTeardown(teardown *v1.Teardown) {
	// try to simulate reality by making things take a little longer. Sleep for a bit...
	time.Sleep(2 * time.Second)

	// change state to running
	teardown.State = v1.TeardownStateInProgress

	systemRecord := backend.getSystemRecordForTeardown(teardown.ID)
	// run service builds
	for sp, s := range systemRecord.system.Services {
		s.State = v1.ServiceStateDeleting

		systemRecord.system.Services[sp] = s
	}

	// sleep
	fmt.Printf("Mock: Running teardown %s. Sleeping for 7 seconds\n", teardown.ID)
	time.Sleep(7 * time.Second)

	systemRecord.system.Services = nil
	teardown.State = v1.TeardownStateSucceeded
}
