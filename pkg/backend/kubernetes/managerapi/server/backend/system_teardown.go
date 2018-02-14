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

func (kb *KubernetesBackend) TearDownSystem(systemID types.SystemID) (types.SystemTeardownID, error) {
	systemTeardown, err := getSystemTeardown(systemID)
	if err != nil {
		return "", err
	}

	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	result, err := kb.latticeClient.LatticeV1().SystemTeardowns(namespace).Create(systemTeardown)
	if err != nil {
		return "", err
	}

	return types.SystemTeardownID(result.Name), err
}

func getSystemTeardown(id types.SystemID) (*latticev1.SystemTeardown, error) {
	labels := map[string]string{
		kubeconstants.LatticeNamespaceLabel: string(id),
	}

	sysT := &latticev1.SystemTeardown{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: latticev1.SystemTeardownSpec{},
		Status: latticev1.SystemTeardownStatus{
			State: latticev1.SystemTeardownStatePending,
		},
	}

	return sysT, nil
}

func (kb *KubernetesBackend) GetSystemTeardown(
	systemID types.SystemID,
	teardownID types.SystemTeardownID,
) (*types.SystemTeardown, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	result, err := kb.latticeClient.LatticeV1().SystemTeardowns(namespace).Get(string(teardownID), metav1.GetOptions{})
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

func (kb *KubernetesBackend) ListSystemTeardowns(systemID types.SystemID) ([]types.SystemTeardown, error) {
	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	result, err := kb.latticeClient.LatticeV1().SystemTeardowns(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var teardowns []types.SystemTeardown
	for _, b := range result.Items {
		teardowns = append(teardowns, types.SystemTeardown{
			ID:    types.SystemTeardownID(b.Name),
			State: getSystemTeardownState(b.Status.State),
		})
	}

	return teardowns, nil
}

func getSystemTeardownState(state latticev1.SystemTeardownState) types.SystemTeardownState {
	switch state {
	case latticev1.SystemTeardownStatePending:
		return types.SystemTeardownStatePending
	case latticev1.SystemTeardownStateInProgress:
		return types.SystemTeardownStateInProgress
	case latticev1.SystemTeardownStateSucceeded:
		return types.SystemTeardownStateSucceeded
	case latticev1.SystemTeardownStateFailed:
		return types.SystemTeardownStateFailed
	default:
		panic("unreachable")
	}
}
