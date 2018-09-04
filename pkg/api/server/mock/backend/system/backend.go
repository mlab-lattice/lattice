package system

import (
	"sync"

	serverv1 "github.com/mlab-lattice/lattice/pkg/api/server/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type systemRecord struct {
	system    *v1.System
	builds    map[v1.BuildID]*v1.Build
	deploys   map[v1.DeployID]*v1.Deploy
	teardowns map[v1.TeardownID]*v1.Teardown
	secrets   []v1.Secret
	nodePools []v1.NodePool
	jobs      map[v1.JobID]*v1.Job
}

type Backend struct {
	controller *controller
	registry   map[v1.SystemID]*systemRecord
	sync.Mutex
}

func NewBackend() *Backend {
	b := &Backend{registry: make(map[v1.SystemID]*systemRecord)}
	c := &controller{backend: b}
	b.controller = c
	return b
}

func (b *Backend) Create(systemID v1.SystemID, definitionURL string) (*v1.System, error) {
	b.Lock()
	defer b.Unlock()

	if _, exists := b.registry[systemID]; exists {
		return nil, v1.NewSystemAlreadyExistsError()
	}

	record := &systemRecord{
		system: &v1.System{
			ID:            systemID,
			State:         v1.SystemStatePending,
			DefinitionURL: definitionURL,
		},
		builds:    make(map[v1.BuildID]*v1.Build),
		deploys:   make(map[v1.DeployID]*v1.Deploy),
		teardowns: make(map[v1.TeardownID]*v1.Teardown),
		jobs:      make(map[v1.JobID]*v1.Job),
		secrets:   []v1.Secret{},
		nodePools: []v1.NodePool{},
	}

	b.registry[systemID] = record
	b.controller.CreateSystem(record)

	system := new(v1.System)
	*system = *record.system
	return system, nil
}

func (b *Backend) List() ([]v1.System, error) {
	b.Lock()
	defer b.Unlock()

	var systems []v1.System
	for _, v := range b.registry {
		systems = append(systems, *v.system)
	}

	return systems, nil
}

func (b *Backend) Get(systemID v1.SystemID) (*v1.System, error) {
	b.Lock()
	defer b.Unlock()

	record, err := b.systemRecord(systemID)
	if err != nil {
		return nil, err
	}

	system := new(v1.System)
	*system = *record.system
	return system, nil
}

func (b *Backend) Delete(systemID v1.SystemID) error {
	b.Lock()
	defer b.Unlock()

	_, err := b.systemRecord(systemID)
	if err != nil {
		return err
	}

	delete(b.registry, systemID)
	return nil
}

func (b *Backend) Builds(id v1.SystemID) serverv1.SystemBuildBackend {
	return &BuildBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Deploys(id v1.SystemID) serverv1.SystemDeployBackend {
	return &DeployBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Jobs(id v1.SystemID) serverv1.SystemJobBackend {
	return &JobBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) NodePools(id v1.SystemID) serverv1.SystemNodePoolBackend {
	return &NodePoolBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Secrets(id v1.SystemID) serverv1.SystemSecretBackend {
	return &SecretBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Services(id v1.SystemID) serverv1.SystemServiceBackend {
	return &ServiceBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Teardowns(id v1.SystemID) serverv1.SystemTeardownBackend {
	return &TeardownBackend{
		backend:  b,
		systemID: id,
	}
}

// helpers
func (b *Backend) systemRecord(systemID v1.SystemID) (*systemRecord, error) {
	systemRecord, ok := b.registry[systemID]
	if !ok {
		return nil, v1.NewInvalidSystemIDError()
	}

	return systemRecord, nil
}
