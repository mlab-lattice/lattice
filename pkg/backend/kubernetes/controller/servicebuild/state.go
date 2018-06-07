package servicebuild

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

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

	successfulComponentBuilds map[string]*latticev1.ContainerBuild
	activeComponentBuilds     map[string]*latticev1.ContainerBuild
	failedComponentBuilds     map[string]*latticev1.ContainerBuild
	needsNewComponentBuilds   []string

	// Maps a component's name to the Name of the Definition that's responsible for it
	componentBuilds map[string]string

	// Maps a Definition.Name to its Definition.Status
	componentBuildStatuses map[string]latticev1.ContainerBuildStatus
}

func (c *Controller) calculateState(build *latticev1.ServiceBuild) (stateInfo, error) {
	successfulComponentBuilds := map[string]*latticev1.ContainerBuild{}
	activeComponentBuilds := map[string]*latticev1.ContainerBuild{}
	failedComponentBuilds := map[string]*latticev1.ContainerBuild{}
	var needsNewComponentBuilds []string

	componentBuilds := map[string]string{}
	componentBuildStatuses := map[string]latticev1.ContainerBuildStatus{}

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

			// If the Definition wasn't in the cache, double check with the API.
			componentBuild, err = c.latticeClient.LatticeV1().ComponentBuilds(build.Namespace).Get(componentBuildName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					err := fmt.Errorf(
						"%v has component build %v for %v, but component build does not exist",
						build.Description(c.namespacePrefix),
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
		case latticev1.ContainerBuildStatePending, latticev1.ContainerBuildStateQueued, latticev1.ContainerBuildStateRunning:
			activeComponentBuilds[component] = componentBuild
		case latticev1.ContainerBuildStateFailed:
			failedComponentBuilds[component] = componentBuild
		case latticev1.ContainerBuildStateSucceeded:
			successfulComponentBuilds[component] = componentBuild
		default:
			// FIXME: send warn event
			err := fmt.Errorf(
				"%v has unexpected state %v",
				componentBuild.Description(c.namespacePrefix),
				componentBuild.Status.State,
			)
			return stateInfo{}, err
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
