package servicebuild

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type state string

const (
	stateHasFailedCBuilds                 state = "has-failed-cbuilds"
	stateHasOnlyRunningOrSucceededCBuilds state = "has-only-succeeded-or-running-cbuilds"
	stateNoFailuresNeedsNewCBuilds        state = "no-failures-needs-new-cbuilds"
	stateAllCBuildsSucceeded              state = "all-cbuilds-succeeded"
)

type stateInfo struct {
	state state

	successfulComponentBuilds map[string]*crv1.ComponentBuild
	activeComponentBuilds     map[string]*crv1.ComponentBuild
	failedComponentBuilds     map[string]*crv1.ComponentBuild
	needsNewComponentBuilds   []string

	componentBuildStatuses map[string]crv1.ComponentBuildStatus
}

func (c *Controller) calculateState(build *crv1.ServiceBuild) (stateInfo, error) {
	successfulComponentBuilds := map[string]*crv1.ComponentBuild{}
	activeComponentBuilds := map[string]*crv1.ComponentBuild{}
	failedComponentBuilds := map[string]*crv1.ComponentBuild{}
	var needsNewComponentBuilds []string

	componentBuildStatuses := map[string]crv1.ComponentBuildStatus{}

	for component := range build.Spec.Components {
		componentBuildName, ok := build.Status.ComponentBuilds[component]
		if !ok {
			needsNewComponentBuilds = append(needsNewComponentBuilds, component)
			continue
		}

		componentBuild, err := c.latticeClient.LatticeV1().ComponentBuilds(build.Namespace).Get(componentBuildName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				err := fmt.Errorf(
					"ServiceBuild %v/%v has ComponentBuild.Name %v for component %v, but ComponentBuild does not exist",
					build.Namespace,
					build.Name,
					componentBuildName,
					component,
				)
				return stateInfo{}, err
			}

			return stateInfo{}, err
		}

		componentBuildStatuses[componentBuild.Name] = componentBuild.Status

		switch componentBuild.Status.State {
		case crv1.ComponentBuildStatePending, crv1.ComponentBuildStateQueued, crv1.ComponentBuildStateRunning:
			activeComponentBuilds[component] = componentBuild
		case crv1.ComponentBuildStateFailed:
			failedComponentBuilds[component] = componentBuild
		case crv1.ComponentBuildStateSucceeded:
			successfulComponentBuilds[component] = componentBuild
		default:
			// FIXME: send warn event
			return stateInfo{}, fmt.Errorf("ComponentBuild %v/%v has unexpected state %v", componentBuild.Namespace, componentBuild.Name, componentBuild.Status.State)
		}
	}

	stateInfo := stateInfo{
		successfulComponentBuilds: successfulComponentBuilds,
		activeComponentBuilds:     activeComponentBuilds,
		failedComponentBuilds:     failedComponentBuilds,
		needsNewComponentBuilds:   needsNewComponentBuilds,
	}

	if len(failedComponentBuilds) > 0 {
		stateInfo.state = stateHasFailedCBuilds
		return stateInfo, nil
	}

	if len(needsNewComponentBuilds) > 0 {
		stateInfo.state = stateNoFailuresNeedsNewCBuilds
		return stateInfo, nil
	}

	if len(activeComponentBuilds) > 0 {
		stateInfo.state = stateHasOnlyRunningOrSucceededCBuilds
		return stateInfo, nil
	}

	stateInfo.state = stateAllCBuildsSucceeded
	return stateInfo, nil
}
