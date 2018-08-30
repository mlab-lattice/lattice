package system

import (
	"time"

	"fmt"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"github.com/satori/go.uuid"
)

type TeardownBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

func (b *TeardownBackend) Create() (*v1.Teardown, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
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
	go b.runTeardown(b.systemID, teardown)

	return teardown, nil
}

func (b *TeardownBackend) List() ([]v1.Teardown, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	var teardowns []v1.Teardown
	for _, teardown := range record.teardowns {
		teardowns = append(teardowns, *teardown)
	}

	return teardowns, nil

}

func (b *TeardownBackend) Get(id v1.TeardownID) (*v1.Teardown, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	teardown, ok := record.teardowns[id]
	if !ok {
		return nil, v1.NewInvalidTeardownIDError(id)
	}

	return teardown, nil
}

func (b *TeardownBackend) runTeardown(systemID v1.SystemID, teardown *v1.Teardown) {
	// try to simulate reality by making things take a little longer. Sleep for a bit...
	time.Sleep(2 * time.Second)

	// change state to running
	teardown.State = v1.TeardownStateInProgress

	b.backend.registryLock.RLock()
	record, err := b.backend.getSystemRecordLocked(systemID)
	if err != nil {
		b.backend.registryLock.RUnlock()
		fmt.Printf("error getting system %v while running teardown %v: %v failing teardown\n", systemID, teardown.ID, err)
		teardown.State = v1.TeardownStateFailed
		return
	}

	record.recordLock.Lock()
	// run service builds
	for path, s := range record.system.Services {
		s.State = v1.ServiceStateDeleting
		record.system.Services[path] = s
	}
	record.recordLock.Unlock()
	b.backend.registryLock.RUnlock()

	// sleep
	fmt.Printf("Running teardown %s. Sleeping for 7 seconds\n", teardown.ID)
	time.Sleep(7 * time.Second)

	record.recordLock.Lock()
	defer record.recordLock.Unlock()

	record.system.Services = make(map[tree.Path]v1.Service)
	teardown.State = v1.TeardownStateSucceeded
}
