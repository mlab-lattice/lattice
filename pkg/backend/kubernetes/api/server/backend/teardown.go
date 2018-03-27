package backend

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) TearDown(systemID v1.SystemID) (*v1.Teardown, error) {
	// ensure the system exists
	if err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	teardown := newTeardown(systemID)

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	teardown, err := kb.latticeClient.LatticeV1().Teardowns(namespace).Create(teardown)
	if err != nil {
		return nil, err
	}

	externalTeardown, err := transformTeardown(teardown)
	if err != nil {
		return nil, err
	}

	return &externalTeardown, nil
}

func newTeardown(id v1.SystemID) *latticev1.Teardown {
	labels := map[string]string{
		kubeconstants.LatticeNamespaceLabel: string(id),
	}

	return &latticev1.Teardown{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: latticev1.TeardownSpec{},
		Status: latticev1.TeardownStatus{
			State: latticev1.TeardownStatePending,
		},
	}
}

func (kb *KubernetesBackend) ListTeardowns(systemID v1.SystemID) ([]v1.Teardown, error) {
	// ensure the system exists
	if err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	teardowns, err := kb.latticeClient.LatticeV1().Teardowns(namespace).List(metav1.ListOptions{})
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

func (kb *KubernetesBackend) GetTeardown(systemID v1.SystemID, teardownID v1.TeardownID) (*v1.Teardown, error) {
	// ensure the system exists
	if err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	teardown, err := kb.latticeClient.LatticeV1().Teardowns(namespace).Get(string(teardownID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidTeardownIDError(teardownID)
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
