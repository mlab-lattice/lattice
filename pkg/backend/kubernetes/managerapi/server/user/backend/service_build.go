package backend

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListServiceBuilds(id types.SystemID) ([]types.ServiceBuild, error) {
	buildList, err := kb.LatticeClient.LatticeV1().ServiceBuilds(string(id)).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var builds []types.ServiceBuild
	for _, build := range buildList.Items {
		externalServiceBuild, err := transformServiceBuild(build.Namespace, build.Name, &build.Status)
		if err != nil {
			return nil, err
		}

		builds = append(builds, externalServiceBuild)
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetServiceBuild(id types.SystemID, bid types.ServiceBuildID) (*types.ServiceBuild, bool, error) {
	build, exists, err := kb.getInternalServiceBuild(id, bid)
	if err != nil || !exists {
		return nil, exists, err
	}

	externalServiceBuild, err := transformServiceBuild(build.Namespace, build.Name, &build.Status)
	if err != nil {
		return nil, true, err
	}

	return &externalServiceBuild, true, nil
}

func (kb *KubernetesBackend) getInternalServiceBuild(id types.SystemID, bid types.ServiceBuildID) (*crv1.ServiceBuild, bool, error) {
	result, err := kb.LatticeClient.LatticeV1().ServiceBuilds(string(id)).Get(string(bid), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return result, true, nil
}

func transformServiceBuild(namespace, name string, status *crv1.ServiceBuildStatus) (types.ServiceBuild, error) {
	externalBuild := types.ServiceBuild{
		ID:         types.ServiceBuildID(name),
		State:      getServiceBuildState(status.State),
		Components: map[string]types.ComponentBuild{},
	}

	for component, componentBuildName := range status.ComponentBuilds {
		componentBuildStatus, ok := status.ComponentBuildStatuses[componentBuildName]
		if !ok {
			err := fmt.Errorf(
				"ServiceBuild %v/%v has ComponentBuild %v for component %v but does not have its status",
				namespace,
				name,
				componentBuildName,
				component,
			)
			return types.ServiceBuild{}, err
		}

		externalComponentBuild := transformComponentBuild(componentBuildName, componentBuildStatus)
		externalBuild.Components[component] = externalComponentBuild
	}

	return externalBuild, nil
}

func getServiceBuildState(state crv1.ServiceBuildState) types.ServiceBuildState {
	switch state {
	case crv1.ServiceBuildStatePending:
		return constants.ServiceBuildStatePending
	case crv1.ServiceBuildStateRunning:
		return constants.ServiceBuildStateRunning
	case crv1.ServiceBuildStateSucceeded:
		return constants.ServiceBuildStateSucceeded
	case crv1.ServiceBuildStateFailed:
		return constants.ServiceBuildStateFailed
	default:
		panic("unreachable")
	}
}
