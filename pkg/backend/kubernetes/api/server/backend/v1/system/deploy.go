package system

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"
	"github.com/satori/go.uuid"
)

type deployBackend struct {
	backend *Backend
	system  v1.SystemID
}

func (b *deployBackend) CreateFromBuild(id v1.BuildID) (*v1.Deploy, error) {
	// this also ensures the system exists
	build, err := b.backend.Builds(b.system).Get(id)
	if err != nil {
		return nil, err
	}

	deploy := newDeploy(build)
	namespace := b.backend.systemNamespace(b.system)
	deploy, err = b.backend.latticeClient.LatticeV1().Deploys(namespace).Create(deploy)
	if err != nil {
		return nil, err
	}

	externalDeploy, err := transformDeploy(deploy)
	if err != nil {
		return nil, err
	}

	return &externalDeploy, nil
}

func (b *deployBackend) CreateFromVersion(version v1.SystemVersion) (*v1.Deploy, error) {
	// this ensures the system is created as well
	build, err := b.backend.Builds(b.system).Create(version)
	if err != nil {
		return nil, err
	}

	return b.CreateFromBuild(build.ID)
}

func newDeploy(build *v1.Build) *latticev1.Deploy {
	return &latticev1.Deploy{
		ObjectMeta: metav1.ObjectMeta{
			Name: uuid.NewV4().String(),
			Labels: map[string]string{
				latticev1.DeployDefinitionVersionLabelKey: string(build.Version),
				latticev1.BuildIDLabelKey:                 string(build.ID),
			},
		},
		Spec: latticev1.DeploySpec{
			Build: string(build.ID),
		},
	}
}

func (b *deployBackend) List() ([]v1.Deploy, error) {
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)
	deploys, err := b.backend.latticeClient.LatticeV1().Deploys(namespace).List(metav1.ListOptions{})
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

func (b *deployBackend) Get(id v1.DeployID) (*v1.Deploy, error) {
	if _, err := b.backend.ensureSystemCreated(b.system); err != nil {
		return nil, err
	}

	namespace := b.backend.systemNamespace(b.system)
	deploy, err := b.backend.latticeClient.LatticeV1().Deploys(namespace).Get(string(id), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidDeployIDError()
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
		BuildID: v1.BuildID(deploy.Spec.Build),
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
