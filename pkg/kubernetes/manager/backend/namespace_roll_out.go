package backend

import (
	"strings"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) RollOutSystem(ln coretypes.LatticeNamespace, definitionRoot systemtree.Node, v coretypes.SystemVersion) (coretypes.SystemRolloutID, error) {
	bid, err := kb.BuildSystem(ln, definitionRoot, v)
	if err != nil {
		return "", err
	}

	return kb.RollOutSystemBuild(ln, bid)
}

func (kb *KubernetesBackend) RollOutSystemBuild(ln coretypes.LatticeNamespace, bid coretypes.SystemBuildID) (coretypes.SystemRolloutID, error) {
	sysBuild, err := kb.getSystemBuildFromId(ln, bid)
	if err != nil {
		return "", err
	}

	sysRollout, err := getNewSystemRollout(ln, sysBuild)
	if err != nil {
		return "", err
	}

	result := &crv1.SystemRollout{}
	err = kb.LatticeResourceClient.Post().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.SystemRolloutResourcePlural).
		Body(sysRollout).
		Do().
		Into(result)

	return coretypes.SystemRolloutID(result.Name), err
}

func (kb *KubernetesBackend) getSystemBuildFromId(ln coretypes.LatticeNamespace, bid coretypes.SystemBuildID) (*crv1.SystemBuild, error) {
	result := &crv1.SystemBuild{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.SystemBuildResourcePlural).
		Name(string(bid)).
		Do().
		Into(result)

	return result, err
}

func getNewSystemRollout(latticeNamespace coretypes.LatticeNamespace, sysBuild *crv1.SystemBuild) (*crv1.SystemRollout, error) {
	labels := map[string]string{
		constants.LatticeNamespaceLabel:   string(latticeNamespace),
		crv1.SystemRolloutVersionLabelKey: sysBuild.Labels[crv1.SystemBuildVersionLabelKey],
		crv1.SystemRolloutBuildIdLabelKey: sysBuild.Name,
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

func (kb *KubernetesBackend) GetSystemRollout(ln coretypes.LatticeNamespace, rid coretypes.SystemRolloutID) (*coretypes.SystemRollout, bool, error) {
	result := &crv1.SystemRollout{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.SystemRolloutResourcePlural).
		Name(string(rid)).
		Do().
		Into(result)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	// TODO: add this to the query
	if strings.Compare(result.Labels[constants.LatticeNamespaceLabel], string(ln)) != 0 {
		return nil, false, nil
	}

	sb := &coretypes.SystemRollout{
		ID:      rid,
		BuildId: coretypes.SystemBuildID(result.Spec.BuildName),
		State:   getSystemRolloutState(result.Status.State),
	}

	return sb, true, nil
}

func (kb *KubernetesBackend) ListSystemRollouts(ln coretypes.LatticeNamespace) ([]coretypes.SystemRollout, error) {
	result := &crv1.SystemRolloutList{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.SystemRolloutResourcePlural).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	rollouts := []coretypes.SystemRollout{}
	for _, r := range result.Items {
		// TODO: add this to the query
		if strings.Compare(r.Labels[constants.LatticeNamespaceLabel], string(ln)) != 0 {
			continue
		}

		rollouts = append(rollouts, coretypes.SystemRollout{
			ID:      coretypes.SystemRolloutID(r.Name),
			BuildId: coretypes.SystemBuildID(r.Spec.BuildName),
			State:   getSystemRolloutState(r.Status.State),
		})
	}

	return rollouts, nil
}

func getSystemRolloutState(state crv1.SystemRolloutState) coretypes.SystemRolloutState {
	switch state {
	case crv1.SystemRolloutStatePending:
		return coreconstants.SystemRolloutStatePending
	case crv1.SystemRolloutStateAccepted:
		return coreconstants.SystemRolloutStateAccepted
	case crv1.SystemRolloutStateInProgress:
		return coreconstants.SystemRolloutStateInProgress
	case crv1.SystemRolloutStateSucceeded:
		return coreconstants.SystemRolloutStateSucceeded
	case crv1.SystemRolloutStateFailed:
		return coreconstants.SystemRolloutStateFailed
	default:
		panic("unreachable")
	}
}
