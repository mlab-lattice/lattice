package backend

import (
	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
)

func (kb *KubernetesBackend) ListServiceBuilds(ln coretypes.LatticeNamespace) ([]coretypes.ServiceBuild, error) {
	result := &crv1.ServiceBuildList{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.ServiceBuildResourcePlural).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	builds := []coretypes.ServiceBuild{}
	for _, build := range result.Items {
		// FIXME: should add a label to component builds for the lattice namespace
		builds = append(builds, transformServiceBuild(&build))
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetServiceBuild(ln coretypes.LatticeNamespace, bid coretypes.ServiceBuildID) (*coretypes.ServiceBuild, bool, error) {
	build, exists, err := kb.getInternalServiceBuild(ln, bid)
	if err != nil || !exists {
		return nil, exists, err
	}

	coreBuild := transformServiceBuild(build)
	// FIXME: should add a label to component builds for the lattice namespace
	return &coreBuild, true, nil
}

func (kb *KubernetesBackend) getInternalServiceBuild(ln coretypes.LatticeNamespace, bid coretypes.ServiceBuildID) (*crv1.ServiceBuild, bool, error) {
	result := &crv1.ServiceBuild{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.ServiceBuildResourcePlural).
		Name(string(bid)).
		Do().
		Into(result)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return result, true, nil
}

func transformServiceBuild(build *crv1.ServiceBuild) coretypes.ServiceBuild {
	svcb := coretypes.ServiceBuild{
		ID:              coretypes.ServiceBuildID(build.Name),
		State:           getServiceBuildState(build.Status.State),
		ComponentBuilds: map[string]*coretypes.ComponentBuild{},
	}

	for component, cbInfo := range build.Spec.Components {
		if cbInfo.BuildName == nil {
			svcb.ComponentBuilds[component] = nil
			continue
		}
		id := coretypes.ComponentBuildID(*cbInfo.BuildName)

		state := coreconstants.ComponentBuildStatePending
		if cbInfo.BuildState != nil {
			state = getComponentBuildState(*cbInfo.BuildState)
		}

		var failureMessage *string
		if cbInfo.FailureInfo != nil {
			failMessage := getComponentBuildFailureMessage(*cbInfo.FailureInfo)
			failureMessage = &failMessage
		}

		svcb.ComponentBuilds[component] = &coretypes.ComponentBuild{
			ID:                id,
			State:             state,
			LastObservedPhase: cbInfo.LastObservedPhase,
			FailureMessage:    failureMessage,
		}
	}

	return svcb
}

func getServiceBuildState(state crv1.ServiceBuildState) coretypes.ServiceBuildState {
	switch state {
	case crv1.ServiceBuildStatePending:
		return coreconstants.ServiceBuildStatePending
	case crv1.ServiceBuildStateRunning:
		return coreconstants.ServiceBuildStateRunning
	case crv1.ServiceBuildStateSucceeded:
		return coreconstants.ServiceBuildStateSucceeded
	case crv1.ServiceBuildStateFailed:
		return coreconstants.ServiceBuildStateFailed
	default:
		panic("unreachable")
	}
}
