package backend

import (
	"github.com/mlab-lattice/system/pkg/constants"
	kubeconstants "github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListServiceBuilds(ln types.LatticeNamespace) ([]types.ServiceBuild, error) {
	result, err := kb.LatticeClient.V1().ServiceBuilds(kubeconstants.NamespaceLatticeInternal).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	builds := []types.ServiceBuild{}
	for _, build := range result.Items {
		// FIXME: should add a label to component builds for the lattice namespace
		builds = append(builds, transformServiceBuild(&build))
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetServiceBuild(ln types.LatticeNamespace, bid types.ServiceBuildID) (*types.ServiceBuild, bool, error) {
	build, exists, err := kb.getInternalServiceBuild(ln, bid)
	if err != nil || !exists {
		return nil, exists, err
	}

	coreBuild := transformServiceBuild(build)
	// FIXME: should add a label to component builds for the lattice namespace
	return &coreBuild, true, nil
}

func (kb *KubernetesBackend) getInternalServiceBuild(ln types.LatticeNamespace, bid types.ServiceBuildID) (*crv1.ServiceBuild, bool, error) {
	result, err := kb.LatticeClient.V1().ServiceBuilds(kubeconstants.NamespaceLatticeInternal).Get(string(bid), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return result, true, nil
}

func transformServiceBuild(build *crv1.ServiceBuild) types.ServiceBuild {
	svcb := types.ServiceBuild{
		ID:              types.ServiceBuildID(build.Name),
		State:           getServiceBuildState(build.Status.State),
		ComponentBuilds: map[string]*types.ComponentBuild{},
	}

	for component, cbInfo := range build.Spec.Components {
		if cbInfo.BuildName == nil {
			svcb.ComponentBuilds[component] = nil
			continue
		}
		id := types.ComponentBuildID(*cbInfo.BuildName)

		state := constants.ComponentBuildStatePending
		if cbInfo.BuildState != nil {
			state = getComponentBuildState(*cbInfo.BuildState)
		}

		var failureMessage *string
		if cbInfo.FailureInfo != nil {
			failMessage := getComponentBuildFailureMessage(*cbInfo.FailureInfo)
			failureMessage = &failMessage
		}

		svcb.ComponentBuilds[component] = &types.ComponentBuild{
			ID:                id,
			State:             state,
			LastObservedPhase: cbInfo.LastObservedPhase,
			FailureMessage:    failureMessage,
		}
	}

	return svcb
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
