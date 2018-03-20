package backend

import (
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) RollOutSystem(
	systemID types.SystemID,
	definitionRoot tree.Node,
	version types.SystemVersion,
) (types.DeployID, error) {
	bid, err := kb.BuildSystem(systemID, definitionRoot, version)
	if err != nil {
		return "", err
	}

	return kb.RollOutSystemBuild(systemID, bid)
}

func (kb *KubernetesBackend) RollOutSystemBuild(
	systemID types.SystemID,
	buildID types.BuildID,
) (types.DeployID, error) {
	sysBuild, err := kb.getSystemBuildFromID(systemID, buildID)
	if err != nil {
		return "", err
	}

	sysRollout, err := getNewSystemRollout(systemID, sysBuild)
	if err != nil {
		return "", err
	}

	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	result, err := kb.latticeClient.LatticeV1().SystemRollouts(namespace).Create(sysRollout)
	if err != nil {
		return "", err
	}
	return types.DeployID(result.Name), err
}

func (kb *KubernetesBackend) getSystemBuildFromID(
	systemID types.SystemID,
	buildID types.BuildID,
) (*latticev1.SystemBuild, error) {
	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	return kb.latticeClient.LatticeV1().SystemBuilds(namespace).Get(string(buildID), metav1.GetOptions{})
}

func getNewSystemRollout(latticeNamespace types.SystemID, build *latticev1.SystemBuild) (*latticev1.SystemRollout, error) {
	labels := map[string]string{
		kubeconstants.LatticeNamespaceLabel:        string(latticeNamespace),
		kubeconstants.LabelKeySystemRolloutVersion: build.Labels[kubeconstants.LabelKeySystemBuildVersion],
		kubeconstants.LabelKeySystemRolloutBuildID: build.Name,
	}

	sysRollout := &latticev1.SystemRollout{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: latticev1.SystemRolloutSpec{
			BuildName: build.Name,
		},
		Status: latticev1.SystemRolloutStatus{
			State: latticev1.SystemRolloutStatePending,
		},
	}

	return sysRollout, nil
}

func (kb *KubernetesBackend) GetSystemRollout(
	systemID types.SystemID,
	rolloutID types.DeployID,
) (*types.Deploy, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	result, err := kb.latticeClient.LatticeV1().SystemRollouts(namespace).Get(string(rolloutID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	sb := &types.Deploy{
		ID:      rolloutID,
		BuildID: types.BuildID(result.Spec.BuildName),
		State:   getSystemRolloutState(result.Status.State),
	}

	return sb, true, nil
}

func (kb *KubernetesBackend) ListSystemRollouts(systemID types.SystemID) ([]types.Deploy, error) {
	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	result, err := kb.latticeClient.LatticeV1().SystemRollouts(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	rollouts := make([]types.Deploy, 0, len(result.Items))
	for _, r := range result.Items {
		rollouts = append(rollouts, types.Deploy{
			ID:      types.DeployID(r.Name),
			BuildID: types.BuildID(r.Spec.BuildName),
			State:   getSystemRolloutState(r.Status.State),
		})
	}

	return rollouts, nil
}

func getSystemRolloutState(state latticev1.SystemRolloutState) types.DeployState {
	switch state {
	case latticev1.SystemRolloutStatePending:
		return types.DeployStatePending
	case latticev1.SystemRolloutStateAccepted:
		return types.DeployStateAccepted
	case latticev1.SystemRolloutStateInProgress:
		return types.DeployStateInProgress
	case latticev1.SystemRolloutStateSucceeded:
		return types.DeployStateSucceeded
	case latticev1.SystemRolloutStateFailed:
		return types.DeployStateFailed
	default:
		panic("unreachable")
	}
}
