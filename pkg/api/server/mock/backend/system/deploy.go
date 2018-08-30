package system

import (
	"fmt"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/satori/go.uuid"
	"time"
)

type DeployBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

func (b *DeployBackend) CreateFromBuild(id v1.BuildID) (*v1.Deploy, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	// Check that the build exists
	_, err = b.backend.getBuildLocked(b.systemID, id)
	if err != nil {
		return nil, err
	}

	deploy := &v1.Deploy{
		ID:      v1.DeployID(uuid.NewV4().String()),
		BuildID: id,
		State:   v1.DeployStatePending,
	}

	record.deploys[deploy.ID] = deploy

	// run the deploy
	go b.processDeploy(b.systemID, deploy, id)

	return deploy, nil
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
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	var deploys []v1.Deploy
	for _, deploy := range record.deploys {
		deploys = append(deploys, *deploy)
	}

	return deploys, nil
}

func (b *DeployBackend) Get(id v1.DeployID) (*v1.Deploy, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	deploy, ok := record.deploys[id]
	if !ok {
		return nil, v1.NewInvalidDeployIDError(id)
	}

	return deploy, nil
}

func (b *DeployBackend) processDeploy(systemID v1.SystemID, deploy *v1.Deploy, buildID v1.BuildID) {
	record, err := b.backend.getSystemRecord(systemID)
	if err != nil {
		fmt.Printf("system %v deleted while processing deploy %v, failing deploy", systemID, deploy.ID)
		deploy.State = v1.DeployStateFailed
		return
	}

	fmt.Printf("Processing deploy %v...\n", deploy.ID)
	ok := func() bool {
		record.recordLock.Lock()
		defer record.recordLock.Unlock()

		// ensure that there is not other deploy accepted/running
		for _, d := range record.deploys {
			if d.State == v1.DeployStateAccepted || d.State == v1.DeployStateInProgress {
				deploy.State = v1.DeployStateFailed
				fmt.Printf("ERROR: Failing deploy %v. Another deploy for system %v is already accepted/running\n",
					deploy.ID, systemID)

				return false
			}
		}

		return true
	}()
	if !ok {
		return
	}

	fmt.Printf("ACCPETED deploy %v!\n", deploy.ID)
	// set the deployment state to accepted
	deploy.State = v1.DeployStateAccepted
	record.recordLock.Unlock()

	// wait until deploy goes in progress or succeeds
	for {
		build, err := b.backend.Builds(systemID).Get(buildID)
		if err != nil {
			fmt.Printf("got error retrieving build for deploy %v: %v, failing deploy", deploy.ID, err)
			deploy.State = v1.DeployStateFailed
			return
		}

		switch build.State {
		case v1.BuildStateSucceeded:
			deploy.State = v1.DeployStateInProgress
			break

		case v1.BuildStateFailed:
			deploy.State = v1.DeployStateFailed
			return
		}

		time.Sleep(200 * time.Millisecond)
	}

	// sleep for 5 seconds then succeed
	fmt.Printf("Deploy %v is in progress now. Sleeping for 5 seconds", deploy.ID)
	time.Sleep(5 * time.Second)

	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err = b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		fmt.Printf("system %v deleted while processing deploy %v, failing", systemID, deploy.ID)
		deploy.State = v1.DeployStateFailed
		return
	}

	record.recordLock.Lock()
	defer record.recordLock.Unlock()

	build, err := b.backend.getBuildLocked(systemID, buildID)
	if err != nil {
		fmt.Printf("got error retrieving build for deploy %v: %v, failing deploy", deploy.ID, err)
	}

	services := make(map[tree.Path]v1.Service)
	for path := range build.Services {
		services[path] = v1.Service{
			ID:                 v1.ServiceID(uuid.NewV4().String()),
			State:              v1.ServiceStateStable,
			Instances:          []string{uuid.NewV4().String()},
			Path:               path,
			AvailableInstances: 1,
		}
	}

	record.system.Services = services

	deploy.State = v1.DeployStateSucceeded
}
