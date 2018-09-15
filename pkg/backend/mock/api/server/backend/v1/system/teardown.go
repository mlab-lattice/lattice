package system

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/satori/go.uuid"
)

type TeardownBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

func (b *TeardownBackend) Create() (*v1.Teardown, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	teardown := &v1.Teardown{
		ID:    v1.TeardownID(uuid.NewV4().String()),
		State: v1.TeardownStatePending,
	}

	record.Teardowns[teardown.ID] = teardown

	// run the teardown
	b.backend.controller.RunTeardown(teardown, record)

	result := new(v1.Teardown)
	*result = *teardown

	return result, nil
}

func (b *TeardownBackend) List() ([]v1.Teardown, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	var teardowns []v1.Teardown
	for _, teardown := range record.Teardowns {
		teardowns = append(teardowns, *teardown)
	}

	return teardowns, nil

}

func (b *TeardownBackend) Get(id v1.TeardownID) (*v1.Teardown, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	teardown, ok := record.Teardowns[id]
	if !ok {
		return nil, v1.NewInvalidTeardownIDError()
	}

	result := new(v1.Teardown)
	*result = *teardown

	return result, nil
}
