package system

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/satori/go.uuid"
)

type DeployBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

func (b *DeployBackend) CreateFromBuild(id v1.BuildID) (*v1.Deploy, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	_, ok := record.builds[id]
	if !ok {
		return nil, v1.NewInvalidBuildIDError()
	}

	deploy := &v1.Deploy{
		ID:      v1.DeployID(uuid.NewV4().String()),
		BuildID: id,
		State:   v1.DeployStatePending,
	}

	record.deploys[deploy.ID] = deploy

	b.backend.controller.RunDeploy(deploy, record)

	// copy the deploy so we don't return a pointer into the backend
	// so we can release the lock
	result := new(v1.Deploy)
	*result = *deploy

	return result, nil
}

func (b *DeployBackend) CreateFromVersion(v v1.SystemVersion) (*v1.Deploy, error) {
	// this ensures the system is created as well
	build, err := b.backend.Builds(b.systemID).Create(v)
	if err != nil {
		return nil, err
	}

	return b.CreateFromBuild(build.ID)
}

func (b *DeployBackend) List() ([]v1.Deploy, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	var deploys []v1.Deploy
	for _, deploy := range record.deploys {
		deploys = append(deploys, *deploy)
	}

	return deploys, nil
}

func (b *DeployBackend) Get(id v1.DeployID) (*v1.Deploy, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	deploy, ok := record.deploys[id]
	if !ok {
		return nil, v1.NewInvalidDeployIDError()
	}

	// copy the deploy so we don't return a pointer into the backend
	// so we can release the lock
	result := new(v1.Deploy)
	*result = *deploy

	return deploy, nil
}
