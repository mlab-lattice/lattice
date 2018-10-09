package system

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/satori/go.uuid"
)

type DeployBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

func (b *DeployBackend) CreateFromBuild(id v1.BuildID) (*v1.Deploy, error) {
	return b.create(&id, nil, nil)
}

func (b *DeployBackend) CreateFromPath(p tree.Path) (*v1.Deploy, error) {
	return b.create(nil, &p, nil)
}

func (b *DeployBackend) CreateFromVersion(v v1.Version) (*v1.Deploy, error) {
	return b.create(nil, nil, &v)
}

func (b *DeployBackend) create(id *v1.BuildID, p *tree.Path, v *v1.Version) (*v1.Deploy, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	deploy := &v1.Deploy{
		ID:      v1.DeployID(uuid.NewV4().String()),
		Build:   id,
		Path:    p,
		Version: v,
		Status: v1.DeployStatus{
			State: v1.DeployStatePending,
		},
	}

	record.Deploys[deploy.ID] = deploy

	b.backend.controller.RunDeploy(deploy, record)

	// copy the deploy so we don't return a pointer into the backend
	// so we can release the lock
	return deploy.DeepCopy(), nil
}

func (b *DeployBackend) List() ([]v1.Deploy, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	var deploys []v1.Deploy
	for _, deploy := range record.Deploys {
		deploys = append(deploys, *deploy.DeepCopy())
	}

	return deploys, nil
}

func (b *DeployBackend) Get(id v1.DeployID) (*v1.Deploy, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	deploy, ok := record.Deploys[id]
	if !ok {
		return nil, v1.NewInvalidDeployIDError()
	}

	// copy the deploy so we don't return a pointer into the backend
	// so we can release the lock
	return deploy.DeepCopy(), nil
}
