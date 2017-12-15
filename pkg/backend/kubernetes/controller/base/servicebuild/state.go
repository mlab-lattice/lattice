package servicebuild

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type state string

const (
	stateHasFailedComponentBuilds                 state = "has-failed-component-builds"
	stateHasOnlyRunningOrSucceededComponentBuilds state = "has-only-succeeded-or-running-component-builds"
	stateNoFailuresNeedsNewComponentBuilds        state = "no-failures-needs-new-component-builds"
	stateAllComponentBuildsSucceeded              state = "all-component-builds-succeeded"
)

type stateInfo struct {
	state state

	successfulComponentBuilds map[string]*crv1.ComponentBuild
	activeComponentBuilds     map[string]*crv1.ComponentBuild
	failedComponentBuilds     map[string]*crv1.ComponentBuild
	needsNewComponentBuilds   []string

	// Maps a component's name to the Name of the ComponentBuild that's responsible for it
	componentBuilds map[string]string

	// Maps a ComponentBuild.Name to its ComponentBuild.Status
	componentBuildStatuses map[string]crv1.ComponentBuildStatus
}

func (c *Controller) calculateState(build *crv1.ServiceBuild) (stateInfo, error) {
	successfulComponentBuilds := map[string]*crv1.ComponentBuild{}
	activeComponentBuilds := map[string]*crv1.ComponentBuild{}
	failedComponentBuilds := map[string]*crv1.ComponentBuild{}
	var needsNewComponentBuilds []string

	componentBuilds := map[string]string{}
	componentBuildStatuses := map[string]crv1.ComponentBuildStatus{}

	for component := range build.Spec.Components {
		componentBuildName, ok := build.Status.ComponentBuilds[component]
		if !ok {
			needsNewComponentBuilds = append(needsNewComponentBuilds, component)
			continue
		}

		componentBuild, err := c.componentBuildLister.ComponentBuilds(build.Namespace).Get(componentBuildName)
		if err != nil {
			if !errors.IsNotFound(err) {
				return stateInfo{}, err
			}

			// If the ComponentBuild wasn't in the cache, double check with the API.
			componentBuild, err = c.latticeClient.LatticeV1().ComponentBuilds(build.Namespace).Get(componentBuildName, metav1.GetOptions{})
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
		}

		componentBuilds[component] = componentBuild.Name
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

		componentBuilds:        componentBuilds,
		componentBuildStatuses: componentBuildStatuses,
	}

	if len(failedComponentBuilds) > 0 {
		stateInfo.state = stateHasFailedComponentBuilds
		return stateInfo, nil
	}

	if len(needsNewComponentBuilds) > 0 {
		stateInfo.state = stateNoFailuresNeedsNewComponentBuilds
		return stateInfo, nil
	}

	if len(activeComponentBuilds) > 0 {
		stateInfo.state = stateHasOnlyRunningOrSucceededComponentBuilds
		return stateInfo, nil
	}

	stateInfo.state = stateAllComponentBuildsSucceeded
	return stateInfo, nil
}
