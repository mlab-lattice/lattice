package systembuild

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"

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

	successfulServiceBuilds map[tree.NodePath]*crv1.ServiceBuild
	activeServiceBuilds     map[tree.NodePath]*crv1.ServiceBuild
	failedServiceBuilds     map[tree.NodePath]*crv1.ServiceBuild
	needsNewServiceBuilds   []tree.NodePath

	// Maps a service's path to the Name of the ServiceBuild that's responsible for it
	serviceBuilds map[tree.NodePath]string

	// Maps a ServiceBuild.Name to its ServiceBuild.Status
	serviceBuildStatuses map[string]crv1.ServiceBuildStatus
}

func (c *Controller) calculateState(build *crv1.SystemBuild) (stateInfo, error) {
	successfulServiceBuilds := map[tree.NodePath]*crv1.ServiceBuild{}
	activeServiceBuilds := map[tree.NodePath]*crv1.ServiceBuild{}
	failedServiceBuilds := map[tree.NodePath]*crv1.ServiceBuild{}
	var needsNewServiceBuilds []tree.NodePath

	serviceBuilds := map[tree.NodePath]string{}
	serviceBuildStatuses := map[string]crv1.ServiceBuildStatus{}

	for service := range build.Spec.Services {
		var serviceBuild *crv1.ServiceBuild
		serviceBuildName, ok := build.Status.ServiceBuilds[service]
		if !ok {
			// It's possible that the ServiceBuild was created, but prior to updating build.Status there was an error.
			// So first check to see if there's a ServiceBuild with the proper labels.
			// TODO: currently using the cache here. I believe it's possible to have created the ServiceBuild and have
			// this cache lookup miss. That said, I don't think there is any real downside to orphaning a ServiceBuild,
			// so we'll optimize here for assuming we don't hit the unfortunate error timing _and_ lose the race, and
			// if we do we'll orphan a ServiceBuild, which is okay.
			labelSelector := fmt.Sprintf("%v==%v,%v==%v", constants.LabelKeySystemBuildID, build.Name, constants.LabelKeyServicePath, service)
			listOptions := metav1.ListOptions{
				LabelSelector: labelSelector,
			}
			serviceBuilds, err := c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).List(listOptions)
			if err != nil {
				return stateInfo{}, err
			}

			if len(serviceBuilds.Items) == 0 {
				needsNewServiceBuilds = append(needsNewServiceBuilds, service)
				continue
			}

			// It's possible that we lost the race multiple times. Assume that this is what happened if we get
			// multiple ServiceBuilds back. They should all be the same, so just choose the first one.
			serviceBuild = &serviceBuilds.Items[0]
		}

		// If we already knew the name of the ServiceBuild, go find it
		if serviceBuild == nil {
			var err error
			serviceBuild, err = c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).Get(serviceBuildName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					err := fmt.Errorf(
						"SystemBuild %v/%v has ServiceBuild.Name %v for service %v, but ServiceBuild does not exist",
						build.Namespace,
						build.Name,
						serviceBuildName,
						service,
					)
					return stateInfo{}, err
				}

				return stateInfo{}, err
			}
		}

		serviceBuilds[service] = serviceBuild.Name
		serviceBuildStatuses[serviceBuild.Name] = serviceBuild.Status

		switch serviceBuild.Status.State {
		case crv1.ServiceBuildStatePending, crv1.ServiceBuildStateRunning:
			activeServiceBuilds[service] = serviceBuild
		case crv1.ServiceBuildStateFailed:
			failedServiceBuilds[service] = serviceBuild
		case crv1.ServiceBuildStateSucceeded:
			successfulServiceBuilds[service] = serviceBuild
		default:
			// FIXME: send warn event
			return stateInfo{}, fmt.Errorf("ServiceBuild %v/%v has unexpected state %v", serviceBuild.Namespace, serviceBuild.Name, serviceBuild.Status.State)
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
