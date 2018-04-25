package build

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type state string

const (
	stateHasFailedServiceBuilds                 state = "has-failed-service-builds"
	stateHasOnlyRunningOrSucceededServiceBuilds state = "has-only-succeeded-or-running-service-builds"
	stateNoFailuresNeedsNewServiceBuilds        state = "no-failures-needs-new-service-builds"
	stateAllServiceBuildsSucceeded              state = "all-service-builds-succeeded"
)

type stateInfo struct {
	state state

	successfulServiceBuilds map[tree.NodePath]*latticev1.ServiceBuild
	activeServiceBuilds     map[tree.NodePath]*latticev1.ServiceBuild
	failedServiceBuilds     map[tree.NodePath]*latticev1.ServiceBuild
	needsNewServiceBuilds   []tree.NodePath

	// Maps a service's path to the Name of the ServiceBuild that's responsible for it
	serviceBuilds map[tree.NodePath]string

	// Maps a ServiceBuild.Name to its ServiceBuild.Status
	serviceBuildStatuses map[string]latticev1.ServiceBuildStatus
}

func (c *Controller) calculateState(build *latticev1.Build) (stateInfo, error) {
	successfulServiceBuilds := make(map[tree.NodePath]*latticev1.ServiceBuild)
	activeServiceBuilds := make(map[tree.NodePath]*latticev1.ServiceBuild)
	failedServiceBuilds := make(map[tree.NodePath]*latticev1.ServiceBuild)
	var needsNewServiceBuilds []tree.NodePath

	serviceBuilds := make(map[tree.NodePath]string)
	serviceBuildStatuses := make(map[string]latticev1.ServiceBuildStatus)

	for servicePath := range build.Spec.Services {
		var serviceBuild *latticev1.ServiceBuild

		serviceBuildName, ok := build.Status.ServiceBuilds[servicePath]
		if !ok {
			if len(serviceBuilds) == 0 {
				needsNewServiceBuilds = append(needsNewServiceBuilds, servicePath)
				continue
			}
		}

		// if there is already a service build for this path, get it
		serviceBuild, err := c.serviceBuildLister.ServiceBuilds(build.Namespace).Get(serviceBuildName)
		if err != nil {
			if !errors.IsNotFound(err) {
				return stateInfo{}, err
			}

			serviceBuild, err = c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).Get(serviceBuildName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					err := fmt.Errorf(
						"%v has service build %v for %v, but service build does not exist",
						build.Description(c.namespacePrefix),
						serviceBuildName,
						servicePath.String(),
					)
					return stateInfo{}, err
				}

				return stateInfo{}, err
			}
		}

		serviceBuilds[servicePath] = serviceBuild.Name
		serviceBuildStatuses[serviceBuild.Name] = serviceBuild.Status

		switch serviceBuild.Status.State {
		case latticev1.ServiceBuildStatePending, latticev1.ServiceBuildStateRunning:
			activeServiceBuilds[servicePath] = serviceBuild
		case latticev1.ServiceBuildStateFailed:
			failedServiceBuilds[servicePath] = serviceBuild
		case latticev1.ServiceBuildStateSucceeded:
			successfulServiceBuilds[servicePath] = serviceBuild
		default:
			// FIXME: send warn event
			err := fmt.Errorf(
				"%v has unexpected state %v",
				serviceBuild.Description(c.namespacePrefix),
				serviceBuild.Status.State,
			)
			return stateInfo{}, err
		}
	}

	stateInfo := stateInfo{
		successfulServiceBuilds: successfulServiceBuilds,
		activeServiceBuilds:     activeServiceBuilds,
		failedServiceBuilds:     failedServiceBuilds,
		needsNewServiceBuilds:   needsNewServiceBuilds,

		serviceBuilds:        serviceBuilds,
		serviceBuildStatuses: serviceBuildStatuses,
	}

	if len(failedServiceBuilds) > 0 {
		stateInfo.state = stateHasFailedServiceBuilds
		return stateInfo, nil
	}

	if len(needsNewServiceBuilds) > 0 {
		stateInfo.state = stateNoFailuresNeedsNewServiceBuilds
		return stateInfo, nil
	}

	if len(activeServiceBuilds) > 0 {
		stateInfo.state = stateHasOnlyRunningOrSucceededServiceBuilds
		return stateInfo, nil
	}

	stateInfo.state = stateAllServiceBuildsSucceeded
	return stateInfo, nil
}
