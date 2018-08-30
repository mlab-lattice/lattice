package system

import (
	"fmt"
	"sync"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type systemRecord struct {
	system     *v1.System
	builds     map[v1.BuildID]*v1.Build
	deploys    map[v1.DeployID]*v1.Deploy
	teardowns  map[v1.TeardownID]*v1.Teardown
	secrets    []v1.Secret
	nodePools  []v1.NodePool
	jobs       map[v1.JobID]*v1.Job
	recordLock sync.RWMutex
}

// newSystemRecord
func newSystemRecord(system *v1.System) *systemRecord {
	return &systemRecord{
		system:    system,
		builds:    make(map[v1.BuildID]*v1.Build),
		deploys:   make(map[v1.DeployID]*v1.Deploy),
		teardowns: make(map[v1.TeardownID]*v1.Teardown),
		jobs:      make(map[v1.JobID]*v1.Job),
		secrets:   []v1.Secret{},
		nodePools: []v1.NodePool{},
	}
}

type Backend struct {
	registry     map[v1.SystemID]*systemRecord
	registryLock sync.RWMutex
}

func NewBackend() *Backend {
	return &Backend{registry: make(map[v1.SystemID]*systemRecord)}
}

func (b *Backend) Create(systemID v1.SystemID, definitionURL string) (*v1.System, error) {
	b.registryLock.Lock()
	defer b.registryLock.Unlock()

	if _, exists := b.registry[systemID]; exists {
		return nil, v1.NewSystemAlreadyExistsError(systemID)
	}

	// create system
	system := &v1.System{
		ID:            systemID,
		State:         v1.SystemStateStable,
		DefinitionURL: definitionURL,
	}

	// register it with in memory map
	b.registry[systemID] = newSystemRecord(system)
	return system, nil
}

func (b *Backend) List() ([]v1.System, error) {
	b.registryLock.RLock()
	defer b.registryLock.RUnlock()

	var systems []v1.System
	for _, v := range b.registry {
		systems = append(systems, *v.system)
	}

	return systems, nil
}

func (b *Backend) Get(systemID v1.SystemID) (*v1.System, error) {
	systemRecord, err := b.getSystemRecord(systemID)
	if err != nil {
		return nil, err
	}

	return systemRecord.system, nil
}

func (b *Backend) Delete(systemID v1.SystemID) error {
	_, err := b.getSystemRecord(systemID)
	if err != nil {
		return err
	}

	// lock for writing
	b.registryLock.Lock()
	defer b.registryLock.Unlock()

	delete(b.registry, systemID)
	return nil
}

func (b *Backend) Builds(id v1.SystemID) *BuildBackend {
	return &BuildBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Deploys(id v1.SystemID) *DeployBackend {
	return &DeployBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Jobs(id v1.SystemID) *JobBackend {
	return &JobBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) NodePools(id v1.SystemID) *NodePoolBackend {
	return &NodePoolBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Secrets(id v1.SystemID) *SecretBackend {
	return &SecretBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Services(id v1.SystemID) *ServiceBackend {
	return &ServiceBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Teardowns(id v1.SystemID) *TeardownBackend {
	return &TeardownBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) getBuildLocked(systemID v1.SystemID, buildID v1.BuildID) (*v1.Build, error) {
	record, err := b.getSystemRecordLocked(systemID)
	if err != nil {
		return nil, err
	}

	build, ok := record.builds[buildID]
	if !ok {
		return nil, v1.NewInvalidBuildIDError(buildID)
	}

	return build, nil
}

// helpers

func (b *Backend) getSystemRecord(systemID v1.SystemID) (*systemRecord, error) {
	b.registryLock.RLock()
	defer b.registryLock.RUnlock()
	return b.getSystemRecordLocked(systemID)
}

func (b *Backend) getSystemRecordLocked(systemID v1.SystemID) (*systemRecord, error) {
	systemRecord, ok := b.registry[systemID]
	if !ok {
		return nil, v1.NewInvalidSystemIDError(systemID)
	}

	return systemRecord, nil
}

func (b *Backend) getSystemRecordForTeardown(teardownID v1.TeardownID) *systemRecord {
	for _, systemRecord := range b.registry {
		if _, exists := systemRecord.teardowns[teardownID]; exists {
			return systemRecord
		}
	}
	return nil
}

func (b *Backend) runTeardown(teardown *v1.Teardown) {
	// try to simulate reality by making things take a little longer. Sleep for a bit...
	time.Sleep(2 * time.Second)

	// change state to running
	teardown.State = v1.TeardownStateInProgress

	systemRecord := b.getSystemRecordForTeardown(teardown.ID)
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
