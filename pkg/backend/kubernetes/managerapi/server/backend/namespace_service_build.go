package backend

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListServiceBuilds(ln types.LatticeNamespace) ([]types.ServiceBuild, error) {
	buildList, err := kb.LatticeClient.LatticeV1().ServiceBuilds(string(ln)).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var builds []types.ServiceBuild
	for _, build := range buildList.Items {
		externalServiceBuild, err := transformServiceBuild(&build)
		if err != nil {
			return nil, err
		}

		builds = append(builds, externalServiceBuild)
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetServiceBuild(ln types.LatticeNamespace, bid types.ServiceBuildID) (*types.ServiceBuild, bool, error) {
	build, exists, err := kb.getInternalServiceBuild(ln, bid)
	if err != nil || !exists {
		return nil, exists, err
	}

	externalServiceBuild, err := transformServiceBuild(build)
	if err != nil {
		return nil, true, err
	}

	return &externalServiceBuild, true, nil
}

func (kb *KubernetesBackend) getInternalServiceBuild(ln types.LatticeNamespace, bid types.ServiceBuildID) (*crv1.ServiceBuild, bool, error) {
	result, err := kb.LatticeClient.LatticeV1().ServiceBuilds(string(ln)).Get(string(bid), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return result, true, nil
}

func transformServiceBuild(build *crv1.ServiceBuild) (types.ServiceBuild, error) {
	externalBuild := types.ServiceBuild{
		ID:              types.ServiceBuildID(build.Name),
		State:           getServiceBuildState(build.Status.State),
		ComponentBuilds: map[string]*types.ComponentBuild{},
	}

	for component := range build.Spec.Components {
		externalComponentBuild, err := transformServiceBuildComponent(
			build.Name,
			build.Namespace,
			component,
			build.Status.ComponentBuilds,
			build.Status.ComponentBuildStatuses,
		)
		if err != nil {
			return types.ServiceBuild{}, err
		}

		externalBuild.ComponentBuilds[component] = externalComponentBuild
	}

	return externalBuild, nil
}

func transformServiceBuildComponent(
	name, namespace, component string,
	componentBuilds map[string]string,
	componentBuildStatuses map[string]crv1.ComponentBuildStatus,
) (*types.ComponentBuild, error) {
	componentBuildName, ok := componentBuilds[component]
	if !ok {
		return nil, nil
	}

	componentBuildStatus, ok := componentBuildStatuses[componentBuildName]
	if !ok {
		err := fmt.Errorf(
			"ServiceBuild %v/%v has ComponentBuild %v for component %v but does not have its status",
			namespace,
			name,
			componentBuildName,
			component,
		)
		return nil, err
	}

	externalComponentBuild := transformComponentBuild(componentBuildName, componentBuildStatus)
	return &externalComponentBuild, nil
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
