package build

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
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
	successfulServiceBuilds := map[tree.NodePath]*latticev1.ServiceBuild{}
	activeServiceBuilds := map[tree.NodePath]*latticev1.ServiceBuild{}
	failedServiceBuilds := map[tree.NodePath]*latticev1.ServiceBuild{}
	var needsNewServiceBuilds []tree.NodePath

	// FIXME: reuse existing service builds

	serviceBuilds := map[tree.NodePath]string{}
	serviceBuildStatuses := map[string]latticev1.ServiceBuildStatus{}

	for servicePath := range build.Spec.Services {
		var serviceBuild *latticev1.ServiceBuild
		serviceBuildName, ok := build.Status.ServiceBuilds[servicePath]
		if !ok {
			// It's possible that the ServiceBuild was created, but prior to updating build.Status there was an error.
			// So first check to see if there's a ServiceBuild with the proper labels.
			// TODO: currently using the cache here. I believe it's possible to have created the ServiceBuild and have
			// this cache lookup miss. That said, I don't think there is any real downside to orphaning a ServiceBuild,
			// so we'll optimize here for assuming we don't hit the unfortunate error timing _and_ lose the race, and
			// if we do we'll orphan a ServiceBuild, which is okay.
			selector := kubelabels.NewSelector()
			requirement, err := kubelabels.NewRequirement(latticev1.BuildIDLabelKey, selection.Equals, []string{build.Name})
			if err != nil {
				return stateInfo{}, err
			}
			selector = selector.Add(*requirement)

			//requirement, err = kubelabels.NewRequirement(
			//	constants.LabelKeyServicePath,
			//	selection.Equals,
			//	[]string{servicePath.ToDomain()},
			//)
			if err != nil {
				return stateInfo{}, err
			}
			selector = selector.Add(*requirement)

			serviceBuilds, err := c.serviceBuildLister.ServiceBuilds(build.Namespace).List(selector)
			if err != nil {
				return stateInfo{}, err
			}

			if len(serviceBuilds) == 0 {
				needsNewServiceBuilds = append(needsNewServiceBuilds, servicePath)
				continue
			}

			// It's possible that we lost the race multiple times. Assume that this is what happened if we get
			// multiple ServiceBuilds back. They should all be the same, so just choose the first one.
			serviceBuild = serviceBuilds[0]
		}

		// If we already knew the name of the ServiceBuild, go find it
		if serviceBuild == nil {
			var err error
			serviceBuild, err = c.serviceBuildLister.ServiceBuilds(build.Namespace).Get(serviceBuildName)
			if err != nil {
				if !errors.IsNotFound(err) {
					return stateInfo{}, err
				}

				serviceBuild, err = c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).Get(serviceBuildName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						err := fmt.Errorf(
							"Build %v/%v has ServiceBuild.Name %v for servicePath %v, but ServiceBuild does not exist",
							build.Namespace,
							build.Name,
							serviceBuildName,
							servicePath,
						)
						return stateInfo{}, err
					}

					return stateInfo{}, err
				}
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
				"ServiceBuild %v/%v has unexpected state %v",
				serviceBuild.Namespace,
				serviceBuild.Name,
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
