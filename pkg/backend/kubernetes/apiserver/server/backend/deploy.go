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

func (kb *KubernetesBackend) DeployBuild(
	systemID types.SystemID,
	buildID types.BuildID,
) (types.DeployID, error) {
	build, err := kb.getBuildFromID(systemID, buildID)
	if err != nil {
		return "", err
	}

	deploy, err := getNewDeploy(systemID, build)
	if err != nil {
		return "", err
	}

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	result, err := kb.latticeClient.LatticeV1().Deploies(namespace).Create(deploy)
	if err != nil {
		return "", err
	}
	return types.DeployID(result.Name), err
}

func (kb *KubernetesBackend) DeployVersion(
	systemID types.SystemID,
	definitionRoot tree.Node,
	version types.SystemVersion,
) (types.DeployID, error) {
	bid, err := kb.Build(systemID, definitionRoot, version)
	if err != nil {
		return "", err
	}

	return kb.DeployBuild(systemID, bid)
}

func (kb *KubernetesBackend) getBuildFromID(
	systemID types.SystemID,
	buildID types.BuildID,
) (*latticev1.Build, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	return kb.latticeClient.LatticeV1().Builds(namespace).Get(string(buildID), metav1.GetOptions{})
}

func getNewDeploy(latticeNamespace types.SystemID, build *latticev1.Build) (*latticev1.Deploy, error) {
	labels := map[string]string{
		kubeconstants.LatticeNamespaceLabel:        string(latticeNamespace),
		kubeconstants.LabelKeySystemRolloutVersion: build.Labels[kubeconstants.LabelKeySystemBuildVersion],
		kubeconstants.LabelKeySystemRolloutBuildID: build.Name,
	}

	sysRollout := &latticev1.Deploy{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: latticev1.DeploySpec{
			BuildName: build.Name,
		},
		Status: latticev1.DeployStatus{
			State: latticev1.DeployStatePending,
		},
	}

	return sysRollout, nil
}

func (kb *KubernetesBackend) ListDeploys(systemID types.SystemID) ([]types.Deploy, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	result, err := kb.latticeClient.LatticeV1().Deploies(namespace).List(metav1.ListOptions{})
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

func (kb *KubernetesBackend) GetDeploy(
	systemID types.SystemID,
	rolloutID types.DeployID,
) (*types.Deploy, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	result, err := kb.latticeClient.LatticeV1().Deploies(namespace).Get(string(rolloutID), metav1.GetOptions{})
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

func getSystemRolloutState(state latticev1.DeployState) types.DeployState {
	switch state {
	case latticev1.DeployStatePending:
		return types.DeployStatePending
	case latticev1.DeployStateAccepted:
		return types.DeployStateAccepted
	case latticev1.DeployStateInProgress:
		return types.DeployStateInProgress
	case latticev1.DeployStateSucceeded:
		return types.DeployStateSucceeded
	case latticev1.DeployStateFailed:
		return types.DeployStateFailed
	default:
		panic("unreachable")
	}
}
