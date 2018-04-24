package backend

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"
	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) DeployBuild(systemID v1.SystemID, buildID v1.BuildID) (*v1.Deploy, error) {
	// this ensures the system is created as well
	build, err := kb.GetBuild(systemID, buildID)
	if err != nil {
		return nil, err
	}

	deploy := newDeploy(build)
	namespace := kb.systemNamespace(systemID)
	deploy, err = kb.latticeClient.LatticeV1().Deploys(namespace).Create(deploy)
	if err != nil {
		return nil, err
	}

	externalDeploy, err := transformDeploy(deploy)
	if err != nil {
		return nil, err
	}

	return &externalDeploy, nil
}

func (kb *KubernetesBackend) DeployVersion(systemID v1.SystemID, definitionRoot tree.Node, version v1.SystemVersion) (*v1.Deploy, error) {
	// this ensures the system is created as well
	build, err := kb.Build(systemID, definitionRoot, version)
	if err != nil {
		return nil, err
	}

	return kb.DeployBuild(systemID, build.ID)
}

func newDeploy(build *v1.Build) *latticev1.Deploy {
	labels := map[string]string{
		kubeconstants.LabelKeySystemRolloutVersion: string(build.Version),
		kubeconstants.LabelKeySystemRolloutBuildID: string(build.ID),
	}

	return &latticev1.Deploy{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: latticev1.DeploySpec{
			BuildName: string(build.ID),
		},
		Status: latticev1.DeployStatus{
			State: latticev1.DeployStatePending,
		},
	}
}

func (kb *KubernetesBackend) ListDeploys(systemID v1.SystemID) ([]v1.Deploy, error) {
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	deploys, err := kb.latticeClient.LatticeV1().Deploys(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// need to actually allocate the slice here so we return a slice instead of nil
	// if deploys.Items is empty
	externalDeploys := make([]v1.Deploy, 0)
	for _, deploy := range deploys.Items {
		externalDeploy, err := transformDeploy(&deploy)
		if err != nil {
			return nil, err
		}

		externalDeploys = append(externalDeploys, externalDeploy)
	}

	return externalDeploys, nil
}

func (kb *KubernetesBackend) GetDeploy(systemID v1.SystemID, deployID v1.DeployID) (*v1.Deploy, error) {
	if _, err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kb.systemNamespace(systemID)
	deploy, err := kb.latticeClient.LatticeV1().Deploys(namespace).Get(string(deployID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidDeployIDError(deployID)
		}

		return nil, err
	}

	externalDeploy, err := transformDeploy(deploy)
	if err != nil {
		return nil, err
	}

	return &externalDeploy, nil
}

func transformDeploy(deploy *latticev1.Deploy) (v1.Deploy, error) {
	state, err := getDeployState(deploy.Status.State)
	if err != nil {
		return v1.Deploy{}, err
	}

	externalDeploy := v1.Deploy{
		ID:      v1.DeployID(deploy.Name),
		BuildID: v1.BuildID(deploy.Spec.BuildName),
		State:   state,
	}
	return externalDeploy, nil
}

func getDeployState(state latticev1.DeployState) (v1.DeployState, error) {
	switch state {
	case latticev1.DeployStatePending:
		return v1.DeployStatePending, nil
	case latticev1.DeployStateAccepted:
		return v1.DeployStateAccepted, nil
	case latticev1.DeployStateInProgress:
		return v1.DeployStateInProgress, nil
	case latticev1.DeployStateSucceeded:
		return v1.DeployStateSucceeded, nil
	case latticev1.DeployStateFailed:
		return v1.DeployStateFailed, nil
	default:
		return "", fmt.Errorf("invalid deploy state: %v", state)
	}
}
