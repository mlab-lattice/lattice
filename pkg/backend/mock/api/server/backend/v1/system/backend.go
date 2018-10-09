package system

import (
	"fmt"
	"time"

	backendv1 "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/controller"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/registry"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	timeutil "github.com/mlab-lattice/lattice/pkg/util/time"
)

type Backend struct {
	registry   *registry.Registry
	controller *controller.Controller
}

func NewBackend(componentResolver resolver.Interface) *Backend {
	r := registry.New()
	c := controller.New(r, componentResolver)
	return &Backend{
		registry:   r,
		controller: c,
	}
}

func (b *Backend) Define(systemID v1.SystemID, definitionURL string) (*v1.System, error) {
	b.registry.Lock()
	defer b.registry.Unlock()

	if _, exists := b.registry.Systems[systemID]; exists {
		return nil, v1.NewSystemAlreadyExistsError()
	}

	record := &registry.SystemRecord{
		System: &v1.System{
			ID:            systemID,
			DefinitionURL: definitionURL,

			Status: v1.SystemStatus{
				State: v1.SystemStatePending,

				CreationTimestamp: *timeutil.New(time.Now()),
			},
		},
		Definition: resolver.NewResolutionTree(),

		Builds: make(map[v1.BuildID]*registry.BuildInfo),

		Deploys: make(map[v1.DeployID]*v1.Deploy),

		Jobs: make(map[v1.JobID]*registry.JobInfo),

		Secrets: make(map[tree.PathSubcomponent]*v1.Secret),

		Services:     make(map[v1.ServiceID]*registry.ServiceInfo),
		ServicePaths: make(map[tree.Path]v1.ServiceID),

		NodePools: make(map[tree.PathSubcomponent]*v1.NodePool),

		Teardowns: make(map[v1.TeardownID]*v1.Teardown),
	}

	b.registry.Systems[systemID] = record
	b.controller.CreateSystem(record)

	return record.System.DeepCopy(), nil
}

func (b *Backend) List() ([]v1.System, error) {
	b.registry.Lock()
	defer b.registry.Unlock()

	var systems []v1.System
	for _, s := range b.registry.Systems {
		systems = append(systems, *s.System.DeepCopy())
	}

	return systems, nil
}

func (b *Backend) Get(systemID v1.SystemID) (*v1.System, error) {
	b.registry.Lock()
	defer b.registry.Unlock()

	record, err := b.systemRecord(systemID)
	if err != nil {
		return nil, err
	}

	return record.System.DeepCopy(), nil
}

func (b *Backend) Delete(systemID v1.SystemID) error {
	b.registry.Lock()
	defer b.registry.Unlock()

	_, err := b.systemRecord(systemID)
	if err != nil {
		return err
	}

	delete(b.registry.Systems, systemID)
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
		backend: b,
		system:  id,
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
func (b *Backend) systemRecordInitialized(id v1.SystemID) (*registry.SystemRecord, error) {
	record, err := b.systemRecord(id)
	if err != nil {
		return nil, err
	}

	switch record.System.Status.State {
	case v1.SystemStateDeleting:
		return record, v1.NewSystemDeletingError()
	case v1.SystemStateFailed:
		return record, v1.NewSystemFailedError()
	case v1.SystemStatePending:
		return record, v1.NewSystemPendingError()
	case v1.SystemStateStable, v1.SystemStateDegraded, v1.SystemStateScaling, v1.SystemStateUpdating:
		return record, nil
	default:
		return nil, fmt.Errorf("invalid system state: %v", record.System.Status.State)
	}
}

func (b *Backend) systemRecord(id v1.SystemID) (*registry.SystemRecord, error) {
	record, ok := b.registry.Systems[id]
	if !ok {
		return nil, v1.NewInvalidSystemIDError()
	}

	return record, nil
}
