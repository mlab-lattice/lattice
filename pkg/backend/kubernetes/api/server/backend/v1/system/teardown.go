package system

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/util/time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

type teardownBackend struct {
	backend *Backend
	system  v1.SystemID
}

func (b *teardownBackend) namespace() string {
	return b.backend.systemNamespace(b.system)
}

func (b *teardownBackend) Create() (*v1.Teardown, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	teardown := &latticev1.Teardown{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewV4().String(),
		},
		Spec: latticev1.TeardownSpec{},
	}

	teardown, err := b.backend.latticeClient.LatticeV1().Teardowns(b.namespace()).Create(teardown)
	if err != nil {
		return nil, err
	}

	externalTeardown, err := transformTeardown(teardown)
	if err != nil {
		return nil, err
	}

	return &externalTeardown, nil
}

func (b *teardownBackend) List() ([]v1.Teardown, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	teardowns, err := b.backend.latticeClient.LatticeV1().Teardowns(b.namespace()).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	externalTeardowns := make([]v1.Teardown, len(teardowns.Items))
	for i := 0; i < len(teardowns.Items); i++ {
		externalTeardown, err := transformTeardown(&teardowns.Items[i])
		if err != nil {
			return nil, err
		}

		externalTeardowns[i] = externalTeardown
	}

	return externalTeardowns, nil
}

func (b *teardownBackend) Get(id v1.TeardownID) (*v1.Teardown, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	teardown, err := b.backend.latticeClient.LatticeV1().Teardowns(b.namespace()).Get(string(id), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidTeardownIDError()
		}

		return nil, err
	}

	externalTeardown, err := transformTeardown(teardown)
	if err != nil {
		return nil, err
	}

	return &externalTeardown, nil
}

func transformTeardown(teardown *latticev1.Teardown) (v1.Teardown, error) {
	state, err := getTeardownState(teardown.Status.State)
	if err != nil {
		return v1.Teardown{}, err
	}

	var startTimestamp *time.Time
	if teardown.Status.StartTimestamp != nil {
		startTimestamp = time.New(teardown.Status.StartTimestamp.Time)
	}

	var completionTimestamp *time.Time
	if teardown.Status.CompletionTimestamp != nil {
		completionTimestamp = time.New(teardown.Status.CompletionTimestamp.Time)
	}

	externalTeardown := v1.Teardown{
		ID: v1.TeardownID(teardown.Name),

		Status: v1.TeardownStatus{
			State:   state,
			Message: teardown.Status.Message,

			StartTimestamp:      startTimestamp,
			CompletionTimestamp: completionTimestamp,
		},
	}

	return externalTeardown, nil
}

func getTeardownState(state latticev1.TeardownState) (v1.TeardownState, error) {
	switch state {
	case latticev1.TeardownStatePending:
		return v1.TeardownStatePending, nil
	case latticev1.TeardownStateInProgress:
		return v1.TeardownStateInProgress, nil
	case latticev1.TeardownStateSucceeded:
		return v1.TeardownStateSucceeded, nil
	case latticev1.TeardownStateFailed:
		return v1.TeardownStateFailed, nil
	default:
		return "", fmt.Errorf("invalid teardown state: %v", state)
	}
}
