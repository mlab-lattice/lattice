package backend

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListServiceBuilds(systemID v1.SystemID) ([]v1.ServiceBuild, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	buildList, err := kb.latticeClient.LatticeV1().ServiceBuilds(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var builds []v1.ServiceBuild
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
	systemID v1.SystemID,
	buildID v1.ServiceBuildID,
) (*v1.ServiceBuild, bool, error) {
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
	systemID v1.SystemID, buildID v1.ServiceBuildID,
) (*latticev1.ServiceBuild, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	result, err := kb.latticeClient.LatticeV1().ServiceBuilds(namespace).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return result, true, nil
}

func transformServiceBuild(namespace, name string, status *latticev1.ServiceBuildStatus) (v1.ServiceBuild, error) {
	externalBuild := v1.ServiceBuild{
		ID:         v1.ServiceBuildID(name),
		State:      getServiceBuildState(status.State),
		Components: map[string]v1.ComponentBuild{},
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
			return v1.ServiceBuild{}, err
		}

		externalComponentBuild := transformComponentBuild(componentBuildName, componentBuildStatus)
		externalBuild.Components[component] = externalComponentBuild
	}

	return externalBuild, nil
}

func getServiceBuildState(state latticev1.ServiceBuildState) v1.ServiceBuildState {
	switch state {
	case latticev1.ServiceBuildStatePending:
		return v1.ServiceBuildStatePending
	case latticev1.ServiceBuildStateRunning:
		return v1.ServiceBuildStateRunning
	case latticev1.ServiceBuildStateSucceeded:
		return v1.ServiceBuildStateSucceeded
	case latticev1.ServiceBuildStateFailed:
		return v1.ServiceBuildStateFailed
	default:
		panic("unreachable")
	}
}
