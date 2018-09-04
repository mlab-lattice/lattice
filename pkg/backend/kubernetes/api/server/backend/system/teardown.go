package system

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

type teardownBackend struct {
	backend *Backend
	system  v1.SystemID
}

func (b *teardownBackend) Create() (*v1.Teardown, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	teardown := newTeardown()

	namespace := b.backend.systemNamespace(b.system)
	teardown, err := b.backend.latticeClient.LatticeV1().Teardowns(namespace).Create(teardown)
	if err != nil {
		return nil, err
	}

	externalTeardown, err := transformTeardown(teardown)
	if err != nil {
		return nil, err
	}

	return &externalTeardown, nil
}

func newTeardown() *latticev1.Teardown {
	return &latticev1.Teardown{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewV4().String(),
		},
		Spec: latticev1.TeardownSpec{},
	}
}

func (b *teardownBackend) List() ([]v1.Teardown, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)
	teardowns, err := b.backend.latticeClient.LatticeV1().Teardowns(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalTeardowns []v1.Teardown
	for _, teardown := range teardowns.Items {
		externalTeardown, err := transformTeardown(&teardown)
		if err != nil {
			return nil, err
		}

		externalTeardowns = append(externalTeardowns, externalTeardown)
	}

	return externalTeardowns, nil
}

func (b *teardownBackend) Get(id v1.TeardownID) (*v1.Teardown, error) {
	// ensure the system exists
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)
	teardown, err := b.backend.latticeClient.LatticeV1().Teardowns(namespace).Get(string(id), metav1.GetOptions{})
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

	externalTeardown := v1.Teardown{
		ID:    v1.TeardownID(teardown.Name),
		State: state,
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
