package backend

import (
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) TearDown(systemID types.SystemID) (types.TeardownID, error) {
	systemTeardown, err := getSystemTeardown(systemID)
	if err != nil {
		return "", err
	}

	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	result, err := kb.latticeClient.LatticeV1().Teardowns(namespace).Create(systemTeardown)
	if err != nil {
		return "", err
	}

	return types.TeardownID(result.Name), err
}

func getSystemTeardown(id types.SystemID) (*latticev1.Teardown, error) {
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

func (kb *KubernetesBackend) ListTeardowns(systemID types.SystemID) ([]types.SystemTeardown, error) {
	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	result, err := kb.latticeClient.LatticeV1().Teardowns(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var teardowns []types.SystemTeardown
	for _, b := range result.Items {
		teardowns = append(teardowns, types.SystemTeardown{
			ID:    types.TeardownID(b.Name),
			State: getSystemTeardownState(b.Status.State),
		})
	}

	return teardowns, nil
}

func (kb *KubernetesBackend) GetTeardown(
	systemID types.SystemID,
	teardownID types.TeardownID,
) (*types.SystemTeardown, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	result, err := kb.latticeClient.LatticeV1().Teardowns(namespace).Get(string(teardownID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	sb := &types.SystemTeardown{
		ID:    teardownID,
		State: getSystemTeardownState(result.Status.State),
	}

	return sb, true, nil
}

func getSystemTeardownState(state latticev1.TeardownState) types.TeardownState {
	switch state {
	case latticev1.TeardownStatePending:
		return types.TeardownStatePending
	case latticev1.TeardownStateInProgress:
		return types.TeardownStateInProgress
	case latticev1.TeardownStateSucceeded:
		return types.TeardownStateSucceeded
	case latticev1.TeardownStateFailed:
		return types.TeardownStateFailed
	default:
		panic("unreachable")
	}
}
