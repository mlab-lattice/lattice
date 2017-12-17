package backend

import (
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) RollOutSystem(id types.SystemID, definitionRoot tree.Node, v types.SystemVersion) (types.SystemRolloutID, error) {
	bid, err := kb.BuildSystem(id, definitionRoot, v)
	if err != nil {
		return "", err
	}

	return kb.RollOutSystemBuild(id, bid)
}

func (kb *KubernetesBackend) RollOutSystemBuild(id types.SystemID, bid types.SystemBuildID) (types.SystemRolloutID, error) {
	sysBuild, err := kb.getSystemBuildFromID(id, bid)
	if err != nil {
		return "", err
	}

	sysRollout, err := getNewSystemRollout(id, sysBuild)
	if err != nil {
		return "", err
	}

	namespace := kubeutil.SystemNamespace(kb.ClusterID, id)
	result, err := kb.LatticeClient.LatticeV1().SystemRollouts(namespace).Create(sysRollout)
	if err != nil {
		return "", err
	}
	return types.SystemRolloutID(result.Name), err
}

func (kb *KubernetesBackend) getSystemBuildFromID(id types.SystemID, bid types.SystemBuildID) (*crv1.SystemBuild, error) {
	return kb.LatticeClient.LatticeV1().SystemBuilds(string(id)).Get(string(bid), metav1.GetOptions{})
}

func getNewSystemRollout(latticeNamespace types.SystemID, sysBuild *crv1.SystemBuild) (*crv1.SystemRollout, error) {
	labels := map[string]string{
		kubeconstants.LatticeNamespaceLabel:        string(latticeNamespace),
		kubeconstants.LabelKeySystemRolloutVersion: sysBuild.Labels[kubeconstants.LabelKeySystemBuildVersion],
		kubeconstants.LabelKeySystemRolloutBuildID: sysBuild.Name,
	}

	sysRollout := &crv1.SystemRollout{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: crv1.SystemRolloutSpec{
			BuildName: sysBuild.Name,
		},
		Status: crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStatePending,
		},
	}

	return sysRollout, nil
}

func (kb *KubernetesBackend) GetSystemRollout(id types.SystemID, rid types.SystemRolloutID) (*types.SystemRollout, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.ClusterID, id)
	result, err := kb.LatticeClient.LatticeV1().SystemRollouts(namespace).Get(string(rid), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	sb := &types.SystemRollout{
		ID:      rid,
		BuildID: types.SystemBuildID(result.Spec.BuildName),
		State:   getSystemRolloutState(result.Status.State),
	}

	return sb, true, nil
}

func (kb *KubernetesBackend) ListSystemRollouts(id types.SystemID) ([]types.SystemRollout, error) {
	namespace := kubeutil.SystemNamespace(kb.ClusterID, id)
	result, err := kb.LatticeClient.LatticeV1().SystemRollouts(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	rollouts := []types.SystemRollout{}
	for _, r := range result.Items {
		rollouts = append(rollouts, types.SystemRollout{
			ID:      types.SystemRolloutID(r.Name),
			BuildID: types.SystemBuildID(r.Spec.BuildName),
			State:   getSystemRolloutState(r.Status.State),
		})
	}

	return rollouts, nil
}

func getSystemRolloutState(state crv1.SystemRolloutState) types.SystemRolloutState {
	switch state {
	case crv1.SystemRolloutStatePending:
		return types.SystemRolloutStatePending
	case crv1.SystemRolloutStateAccepted:
		return types.SystemRolloutStateAccepted
	case crv1.SystemRolloutStateInProgress:
		return types.SystemRolloutStateInProgress
	case crv1.SystemRolloutStateSucceeded:
		return types.SystemRolloutStateSucceeded
	case crv1.SystemRolloutStateFailed:
		return types.SystemRolloutStateFailed
	default:
		panic("unreachable")
	}
}
