package backend

import (
	"strings"

	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	kubeconstants "github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) RollOutSystem(ln types.LatticeNamespace, definitionRoot tree.Node, v types.SystemVersion) (types.SystemRolloutID, error) {
	bid, err := kb.BuildSystem(ln, definitionRoot, v)
	if err != nil {
		return "", err
	}

	return kb.RollOutSystemBuild(ln, bid)
}

func (kb *KubernetesBackend) RollOutSystemBuild(ln types.LatticeNamespace, bid types.SystemBuildID) (types.SystemRolloutID, error) {
	sysBuild, err := kb.getSystemBuildFromID(ln, bid)
	if err != nil {
		return "", err
	}

	sysRollout, err := getNewSystemRollout(ln, sysBuild)
	if err != nil {
		return "", err
	}

	result, err := kb.LatticeClient.V1().SystemRollouts(kubeconstants.NamespaceLatticeInternal).Create(sysRollout)
	if err != nil {
		return "", err
	}
	return types.SystemRolloutID(result.Name), err
}

func (kb *KubernetesBackend) getSystemBuildFromID(ln types.LatticeNamespace, bid types.SystemBuildID) (*crv1.SystemBuild, error) {
	return kb.LatticeClient.V1().SystemBuilds(kubeconstants.NamespaceLatticeInternal).Get(string(bid), metav1.GetOptions{})
}

func getNewSystemRollout(latticeNamespace types.LatticeNamespace, sysBuild *crv1.SystemBuild) (*crv1.SystemRollout, error) {
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
			LatticeNamespace: latticeNamespace,
			BuildName:        sysBuild.Name,
		},
		Status: crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStatePending,
		},
	}

	return sysRollout, nil
}

func (kb *KubernetesBackend) GetSystemRollout(ln types.LatticeNamespace, rid types.SystemRolloutID) (*types.SystemRollout, bool, error) {
	result, err := kb.LatticeClient.V1().SystemRollouts(kubeconstants.NamespaceLatticeInternal).Get(string(rid), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	// TODO: add this to the query
	if strings.Compare(result.Labels[kubeconstants.LatticeNamespaceLabel], string(ln)) != 0 {
		return nil, false, nil
	}

	sb := &types.SystemRollout{
		ID:      rid,
		BuildID: types.SystemBuildID(result.Spec.BuildName),
		State:   getSystemRolloutState(result.Status.State),
	}

	return sb, true, nil
}

func (kb *KubernetesBackend) ListSystemRollouts(ln types.LatticeNamespace) ([]types.SystemRollout, error) {
	result, err := kb.LatticeClient.V1().SystemRollouts(kubeconstants.NamespaceLatticeInternal).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	rollouts := []types.SystemRollout{}
	for _, r := range result.Items {
		// TODO: add this to the query
		if strings.Compare(r.Labels[kubeconstants.LatticeNamespaceLabel], string(ln)) != 0 {
			continue
		}

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
		return constants.SystemRolloutStatePending
	case crv1.SystemRolloutStateAccepted:
		return constants.SystemRolloutStateAccepted
	case crv1.SystemRolloutStateInProgress:
		return constants.SystemRolloutStateInProgress
	case crv1.SystemRolloutStateSucceeded:
		return constants.SystemRolloutStateSucceeded
	case crv1.SystemRolloutStateFailed:
		return constants.SystemRolloutStateFailed
	default:
		panic("unreachable")
	}
}
