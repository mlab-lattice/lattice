package system

import (
	"sync"

	"fmt"
	backendv1 "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	syncutil "github.com/mlab-lattice/lattice/pkg/util/sync"
)

type systemRecord struct {
	system     *v1.System
	definition *resolver.ComponentTree

	builds map[v1.BuildID]*buildInfo

	deploys map[v1.DeployID]*v1.Deploy

	jobs map[v1.JobID]*v1.Job

	nodePools map[tree.PathSubcomponent]*v1.NodePool

	secrets map[tree.PathSubcomponent]*v1.Secret

	services     map[v1.ServiceID]*serviceInfo
	servicePaths map[tree.Path]v1.ServiceID

	teardowns map[v1.TeardownID]*v1.Teardown
}

type buildInfo struct {
	Build         *v1.Build
	ComponentTree *resolver.ComponentTree
}

type serviceInfo struct {
	Service    *v1.Service
	Definition *definitionv1.Service
}

type Backend struct {
	controller *controller
	registry   map[v1.SystemID]*systemRecord
	sync.Mutex
}

func NewBackend(componentResolver resolver.ComponentResolver) *Backend {
	b := &Backend{registry: make(map[v1.SystemID]*systemRecord)}
	c := &controller{
		backend:           b,
		actions:           syncutil.NewLifecycleActionManager(),
		componentResolver: componentResolver,
	}
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
		definition: resolver.NewComponentTree(),

		builds: make(map[v1.BuildID]*buildInfo),

		deploys: make(map[v1.DeployID]*v1.Deploy),

		jobs: make(map[v1.JobID]*v1.Job),

		secrets: make(map[tree.PathSubcomponent]*v1.Secret),

		services:     make(map[v1.ServiceID]*serviceInfo),
		servicePaths: make(map[tree.Path]v1.ServiceID),

		nodePools: make(map[tree.PathSubcomponent]*v1.NodePool),

		teardowns: make(map[v1.TeardownID]*v1.Teardown),
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

func (b *Backend) Builds(id v1.SystemID) backendv1.SystemBuildBackend {
	return &BuildBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Deploys(id v1.SystemID) backendv1.SystemDeployBackend {
	return &DeployBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Jobs(id v1.SystemID) backendv1.SystemJobBackend {
	return &JobBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) NodePools(id v1.SystemID) backendv1.SystemNodePoolBackend {
	return &NodePoolBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Secrets(id v1.SystemID) backendv1.SystemSecretBackend {
	return &SecretBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Services(id v1.SystemID) backendv1.SystemServiceBackend {
	return &ServiceBackend{
		backend:  b,
		systemID: id,
	}
}

func (b *Backend) Teardowns(id v1.SystemID) backendv1.SystemTeardownBackend {
	return &TeardownBackend{
		backend:  b,
		systemID: id,
	}
}

// helpers
func (b *Backend) systemRecordInitialized(id v1.SystemID) (*systemRecord, error) {
	record, err := b.systemRecord(id)
	if err != nil {
		return nil, err
	}

	switch record.system.State {
	case v1.SystemStateDeleting:
		return record, v1.NewSystemDeletingError()
	case v1.SystemStateFailed:
		return record, v1.NewSystemFailedError()
	case v1.SystemStatePending:
		return record, v1.NewSystemPendingError()
	case v1.SystemStateStable, v1.SystemStateDegraded, v1.SystemStateScaling, v1.SystemStateUpdating:
		return record, nil
	default:
		return nil, fmt.Errorf("invalid system state: %v", record.system.State)
	}
}

func (b *Backend) systemRecord(id v1.SystemID) (*systemRecord, error) {
	record, ok := b.registry[id]
	if !ok {
		return nil, v1.NewInvalidSystemIDError()
	}

	return record, nil
}
