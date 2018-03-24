package backend

import (
	"github.com/mlab-lattice/system/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) TearDown(systemID v1.SystemID) (v1.TeardownID, error) {
	systemTeardown, err := getSystemTeardown(systemID)
	if err != nil {
		return "", err
	}

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	result, err := kb.latticeClient.LatticeV1().Teardowns(namespace).Create(systemTeardown)
	if err != nil {
		return "", err
	}

	return v1.TeardownID(result.Name), err
}

func getSystemTeardown(id v1.SystemID) (*latticev1.Teardown, error) {
	labels := map[string]string{
		kubeconstants.LatticeNamespaceLabel: string(id),
	}

	sysT := &latticev1.Teardown{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: latticev1.TeardownSpec{},
		Status: latticev1.TeardownStatus{
			State: latticev1.TeardownStatePending,
		},
	}

	return sysT, nil
}

func (kb *KubernetesBackend) ListTeardowns(systemID v1.SystemID) ([]v1.SystemTeardown, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	result, err := kb.latticeClient.LatticeV1().Teardowns(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var teardowns []v1.SystemTeardown
	for _, b := range result.Items {
		teardowns = append(teardowns, v1.SystemTeardown{
			ID:    v1.TeardownID(b.Name),
			State: getSystemTeardownState(b.Status.State),
		})
	}

	return teardowns, nil
}

func (kb *KubernetesBackend) GetTeardown(
	systemID v1.SystemID,
	teardownID v1.TeardownID,
) (*v1.SystemTeardown, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	result, err := kb.latticeClient.LatticeV1().Teardowns(namespace).Get(string(teardownID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	sb := &v1.SystemTeardown{
		ID:    teardownID,
		State: getSystemTeardownState(result.Status.State),
	}

	return sb, true, nil
}

func getSystemTeardownState(state latticev1.TeardownState) v1.TeardownState {
	switch state {
	case latticev1.TeardownStatePending:
		return v1.TeardownStatePending
	case latticev1.TeardownStateInProgress:
		return v1.TeardownStateInProgress
	case latticev1.TeardownStateSucceeded:
		return v1.TeardownStateSucceeded
	case latticev1.TeardownStateFailed:
		return v1.TeardownStateFailed
	default:
		panic("unreachable")
	}
}
