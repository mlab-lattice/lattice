package backend

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListServiceBuilds(systemID types.SystemID) ([]types.ServiceBuild, error) {
	namespace := kubeutil.SystemNamespace(kb.ClusterID, systemID)
	buildList, err := kb.LatticeClient.LatticeV1().ServiceBuilds(namespace).List(metav1.ListOptions{})
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

func (kb *KubernetesBackend) GetServiceBuild(
	systemID types.SystemID,
	buildID types.ServiceBuildID,
) (*types.ServiceBuild, bool, error) {
	build, exists, err := kb.getInternalServiceBuild(systemID, buildID)
	if err != nil || !exists {
		return nil, exists, err
	}

	externalServiceBuild, err := transformServiceBuild(build.Namespace, build.Name, &build.Status)
	if err != nil {
		return nil, true, err
	}

	return &externalServiceBuild, true, nil
}

func (kb *KubernetesBackend) getInternalServiceBuild(
	systemID types.SystemID, buildID types.ServiceBuildID,
) (*latticev1.ServiceBuild, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.ClusterID, systemID)
	result, err := kb.LatticeClient.LatticeV1().ServiceBuilds(namespace).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return result, true, nil
}

func transformServiceBuild(namespace, name string, status *latticev1.ServiceBuildStatus) (types.ServiceBuild, error) {
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

func getServiceBuildState(state latticev1.ServiceBuildState) types.ServiceBuildState {
	switch state {
	case latticev1.ServiceBuildStatePending:
		return types.ServiceBuildStatePending
	case latticev1.ServiceBuildStateRunning:
		return types.ServiceBuildStateRunning
	case latticev1.ServiceBuildStateSucceeded:
		return types.ServiceBuildStateSucceeded
	case latticev1.ServiceBuildStateFailed:
		return types.ServiceBuildStateFailed
	default:
		panic("unreachable")
	}
}
