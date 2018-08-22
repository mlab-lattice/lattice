package mock

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"
	"github.com/satori/go.uuid"
)

type Backend struct {
	systemRegistry map[v1.SystemID]*systemRecord
	registryLock   sync.RWMutex
	gitResolver    *git.Resolver
}

func newMockBackend() (*Backend, error) {
	gitResolver, err := git.NewResolver("/tmp/lattice-api-mock", true)

	if err != nil {
		return nil, err
	}
	return &Backend{
		systemRegistry: make(map[v1.SystemID]*systemRecord),
		gitResolver:    gitResolver,
	}, nil
}

type systemRecord struct {
	system     *v1.System
	builds     map[v1.BuildID]*v1.Build
	deploys    map[v1.DeployID]*v1.Deploy
	teardowns  map[v1.TeardownID]*v1.Teardown
	secrets    []v1.Secret
	nodePools  []v1.NodePool
	recordLock sync.RWMutex
}

// newSystemRecord
func newSystemRecord(system *v1.System) *systemRecord {
	return &systemRecord{
		system:    system,
		builds:    make(map[v1.BuildID]*v1.Build),
		deploys:   make(map[v1.DeployID]*v1.Deploy),
		teardowns: make(map[v1.TeardownID]*v1.Teardown),
		secrets:   []v1.Secret{},
		nodePools: []v1.NodePool{},
	}
}

// Systems
func (backend *Backend) CreateSystem(systemID v1.SystemID, definitionURL string) (*v1.System, error) {
	// lock for writing
	backend.registryLock.Lock()
	defer backend.registryLock.Unlock()

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

func (backend *Backend) ListSystems() ([]v1.System, error) {

	systems := []v1.System{}
	for _, v := range backend.systemRegistry {
		systems = append(systems, *v.system)
	}
	return systems, nil
}

func (backend *Backend) GetSystem(systemID v1.SystemID) (*v1.System, error) {
	systemRecord, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	return systemRecord.system, nil
}

func (backend *Backend) DeleteSystem(systemID v1.SystemID) error {

	_, err := backend.getSystemRecord(systemID)
	if err != nil {
		return err
	}

	// lock for writing
	backend.registryLock.Lock()
	defer backend.registryLock.Unlock()

	delete(backend.systemRegistry, systemID)
	return nil
}

// Builds

func (backend *Backend) Build(
	systemID v1.SystemID,
	def *definitionv1.SystemNode,
	ri resolver.ResolutionInfo,
	v v1.SystemVersion) (*v1.Build, error) {

	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	// validate definition URL
	if !backend.gitResolver.IsValidRepositoryURI(record.system.DefinitionURL) {
		return nil, fmt.Errorf("bad url: %v", record.system.DefinitionURL)
	}
	// validate version
	if v != "1.0.0" {
		return nil, &v1.InvalidSystemVersionError{Version: string(v)}
	}

	build := backend.newMockBuild(systemID, v)

	record.builds[build.ID] = build

	// run the build
	go backend.runBuild(build)

	return build, nil
}

func (backend *Backend) ListBuilds(systemID v1.SystemID) ([]v1.Build, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	builds := []v1.Build{}
	for _, build := range record.builds {
		builds = append(builds, *build)
	}
	return builds, nil

}

func (backend *Backend) GetBuild(systemID v1.SystemID, buildID v1.BuildID) (*v1.Build, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()
	build, exists := record.builds[buildID]

	if !exists {
		return nil, v1.NewInvalidBuildIDError(buildID)
	}

	return build, nil
}

func (backend *Backend) BuildLogs(
	systemID v1.SystemID,
	buildID v1.BuildID,
	path tree.Path,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	return nil, nil
}

// Deploys
func (backend *Backend) DeployBuild(systemID v1.SystemID, buildID v1.BuildID) (*v1.Deploy, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}
	// ensure that build exists
	_, err = backend.GetBuild(systemID, buildID)
	if err != nil {
		return nil, err
	}

	record.recordLock.Lock()
	defer record.recordLock.Unlock()

	deploy := &v1.Deploy{
		ID:      v1.DeployID(uuid.NewV4().String()),
		BuildID: buildID,
		State:   v1.DeployStatePending,
	}

	record.deploys[deploy.ID] = deploy

	// run the deploy
	go backend.monitorDeployBuild(deploy, buildID, systemID)

	return deploy, nil
}

func (backend *Backend) DeployVersion(
	systemID v1.SystemID,
	def *definitionv1.SystemNode,
	ri resolver.ResolutionInfo,
	version v1.SystemVersion) (*v1.Deploy, error) {
	// this ensures the system is created as well
	build, err := backend.Build(systemID, def, ri, version)
	if err != nil {
		return nil, err
	}

	return backend.DeployBuild(systemID, build.ID)
}

func (backend *Backend) ListDeploys(systemID v1.SystemID) ([]v1.Deploy, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	deploys := []v1.Deploy{}
	for _, deploy := range record.deploys {
		deploys = append(deploys, *deploy)
	}
	return deploys, nil
}

func (backend *Backend) GetDeploy(systemID v1.SystemID, deployID v1.DeployID) (*v1.Deploy, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	deploy, exists := record.deploys[deployID]

	if !exists {
		return nil, v1.NewInvalidDeployIDError(deployID)
	}

	return deploy, nil
}

// Teardown
func (backend *Backend) TearDown(systemID v1.SystemID) (*v1.Teardown, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.Lock()
	defer record.recordLock.Unlock()

	teardown := &v1.Teardown{
		ID:    v1.TeardownID(uuid.NewV4().String()),
		State: v1.TeardownStatePending,
	}

	record.teardowns[teardown.ID] = teardown
	// run the teardown
	go backend.runTeardown(teardown)
	return teardown, nil
}

func (backend *Backend) ListTeardowns(systemID v1.SystemID) ([]v1.Teardown, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	teardowns := []v1.Teardown{}
	for _, teardown := range record.teardowns {
		teardowns = append(teardowns, *teardown)
	}
	return teardowns, nil
}

func (backend *Backend) GetTeardown(systemID v1.SystemID, teardownID v1.TeardownID) (*v1.Teardown, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	teardown, exists := record.teardowns[teardownID]

	if !exists {
		return nil, v1.NewInvalidTeardownIDError(teardownID)
	}

	return teardown, nil
}

// Services
func (backend *Backend) ListServices(systemID v1.SystemID) ([]v1.Service, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	services := []v1.Service{}
	for _, service := range record.system.Services {
		services = append(services, service)
	}
	return services, nil
}

func (backend *Backend) GetService(systemID v1.SystemID, serviceID v1.ServiceID) (*v1.Service, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	for _, service := range record.system.Services {
		if service.ID == serviceID {
			return &service, nil
		}
	}

	return nil, v1.NewInvalidServiceIDError(serviceID)
}

func (backend *Backend) GetServiceByPath(systemID v1.SystemID, path tree.Path) (*v1.Service, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	service, exists := record.system.Services[path]

	if !exists {
		return nil, v1.NewInvalidServicePathError(path)
	}

	return &service, nil
}

func (backend *Backend) ServiceLogs(
	systemID v1.SystemID,
	serviceID v1.ServiceID,
	sidecar *string,
	instance string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	return nil, nil
}

// Secrets
func (backend *Backend) ListSystemSecrets(systemID v1.SystemID) ([]v1.Secret, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	return record.secrets, nil
}

func (backend *Backend) GetSystemSecret(systemID v1.SystemID, path tree.Path, name string) (*v1.Secret, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	for _, secret := range record.secrets {
		if secret.Path == path && secret.Name == name {
			scrt := &secret
			return scrt, nil
		}
	}

	return nil, v1.NewInvalidSystemSecretError(path, name)
}

func (backend *Backend) SetSystemSecret(systemID v1.SystemID, path tree.Path, name, value string) error {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return err
	}

	record.recordLock.Lock()
	defer record.recordLock.Unlock()

	secret := v1.Secret{
		Path:  path,
		Name:  name,
		Value: value,
	}

	record.secrets = append(record.secrets, secret)

	return nil
}

func (backend *Backend) UnsetSystemSecret(systemID v1.SystemID, path tree.Path, name string) error {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return err
	}

	record.recordLock.Lock()
	defer record.recordLock.Unlock()

	for i, secret := range record.secrets {
		if secret.Path == path && secret.Name == name {
			// delete secret
			record.secrets = append(record.secrets[:i], record.secrets[i+1:]...)

			return nil
		}
	}

	return v1.NewInvalidSystemSecretError(path, name)
}

func (backend *Backend) ListNodePools(systemID v1.SystemID) ([]v1.NodePool, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	return record.nodePools, nil
}

func (backend *Backend) GetNodePool(systemID v1.SystemID, path v1.NodePoolPath) (*v1.NodePool, error) {
	record, err := backend.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}
	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	for _, nodePool := range record.nodePools {
		if nodePool.Path == path.String() {
			np := &nodePool
			return np, nil
		}
	}
	return nil, nil
}

// Jobs
func (backend *Backend) RunJob(systemID v1.SystemID, path tree.Path, command []string,
	environment definitionv1.ContainerEnvironment,
) (*v1.Job, error) {
	// TODO implement RunJob
	return nil, nil
}

func (backend *Backend) ListJobs(v1.SystemID) ([]v1.Job, error) {
	// TODO implement ListJobs
	return nil, nil
}
func (backend *Backend) GetJob(v1.SystemID, v1.JobID) (*v1.Job, error) {
	// TODO implement GetJob
	return nil, nil
}
func (backend *Backend) JobLogs(systemID v1.SystemID, jobID v1.JobID, sidecar *string, logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	// TODO implement JobLogs
	return nil, nil
}

// helpers

func (backend *Backend) getSystemRecord(systemID v1.SystemID) (*systemRecord, error) {
	backend.registryLock.RLock()
	defer backend.registryLock.RUnlock()
	systemRecord, exists := backend.systemRegistry[systemID]

	if !exists {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	return systemRecord, nil
}

func (backend *Backend) getSystemRecordForBuild(buildID v1.BuildID) *systemRecord {
	for _, systemRecord := range backend.systemRegistry {
		if _, exists := systemRecord.builds[buildID]; exists {
			return systemRecord
		}
	}

	return nil
}

func (backend *Backend) newMockBuild(systemID v1.SystemID, v v1.SystemVersion) *v1.Build {
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

func (backend *Backend) runBuild(build *v1.Build) {
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
	backend.finishBuild(build)

}

func (backend *Backend) finishBuild(build *v1.Build) {
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

	// update system services...
	systemRecord := backend.getSystemRecordForBuild(build.ID)
	services := make(map[tree.Path]v1.Service)
	for _, build := range systemRecord.builds {
		for path := range build.Services {
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

func (backend *Backend) monitorDeployBuild(deploy *v1.Deploy, buildID v1.BuildID, systemID v1.SystemID) {

	for i := 0; i <= 200; i++ {
		build, _ := backend.GetBuild(systemID, buildID)
		newState := getDeployStateByBuild(build.State)
		if deploy.State != newState {
			deploy.State = newState
		}

		if deploy.State == v1.DeployStateSucceeded || deploy.State == v1.DeployStateFailed {
			return
		}

		time.Sleep(200 * time.Millisecond)
	}

	// timed out... fail
	deploy.State = v1.DeployStateFailed

}

func getDeployStateByBuild(buildState v1.BuildState) v1.DeployState {
	switch buildState {
	case v1.BuildStatePending:
		return v1.DeployStatePending
	case v1.BuildStateRunning:
		return v1.DeployStateInProgress
	case v1.BuildStateFailed:
		return v1.DeployStateFailed
	case v1.BuildStateSucceeded:
		return v1.DeployStateSucceeded
	default:
		return v1.DeployStateAccepted
	}
}
func (backend *Backend) getSystemRecordForTeardown(teardownID v1.TeardownID) *systemRecord {
	for _, systemRecord := range backend.systemRegistry {
		if _, exists := systemRecord.teardowns[teardownID]; exists {
			return systemRecord
		}
	}
	return nil
}

func (backend *Backend) runTeardown(teardown *v1.Teardown) {
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
