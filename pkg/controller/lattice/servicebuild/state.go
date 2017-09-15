package servicebuild

import (
	"fmt"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

type svcBuildState string

const (
	svcBuildStateHasFailedCBuilds                 svcBuildState = "has-failed-cbuilds"
	svcBuildStateHasOnlyRunningOrSucceededCBuilds svcBuildState = "has-only-succeeded-or-running-cbuilds"
	svcBuildStateNoFailuresNeedsNewCBuilds        svcBuildState = "no-failures-needs-new-cbuilds"
	svcBuildStateAllCBuildsSucceeded              svcBuildState = "all-cbuilds-succeeded"
)

type svcBuildStateInfo struct {
	state svcBuildState

	activeCBuilds  []string
	failedCBuilds  []string
	needsNewCBuild []string
}

func (sbc *ServiceBuildController) calculateState(svcBuild *crv1.ServiceBuild) (*svcBuildStateInfo, error) {
	activeCBuilds := []string{}
	failedCBuilds := []string{}
	needsNewCBuilds := []string{}

	for componentName, cBuildInfo := range svcBuild.Spec.ComponentBuildsInfo {
		cBuild, exists, err := sbc.getComponentBuildFromInfo(&cBuildInfo, svcBuild.Namespace)
		if err != nil {
			return nil, err
		}

		if !exists {
			needsNewCBuilds = append(needsNewCBuilds, componentName)
			continue
		}

		switch cBuild.Status.State {
		case crv1.ComponentBuildStatePending, crv1.ComponentBuildStateQueued, crv1.ComponentBuildStateRunning:
			activeCBuilds = append(activeCBuilds, componentName)
		case crv1.ComponentBuildStateFailed:
			failedCBuilds = append(failedCBuilds, componentName)
		case crv1.ComponentBuildStateSucceeded:
			continue
		default:
			// FIXME: send warn event
			return nil, fmt.Errorf("ComponentBuild %v has unrecognized state %v", cBuild.Name, cBuild.Status.State)
		}
	}

	stateInfo := &svcBuildStateInfo{
		activeCBuilds:  activeCBuilds,
		failedCBuilds:  failedCBuilds,
		needsNewCBuild: needsNewCBuilds,
	}

	if len(failedCBuilds) > 0 {
		stateInfo.state = svcBuildStateHasFailedCBuilds
		return stateInfo, nil
	}

	if len(needsNewCBuilds) > 0 {
		stateInfo.state = svcBuildStateNoFailuresNeedsNewCBuilds
		return stateInfo, nil
	}

	if len(activeCBuilds) > 0 {
		stateInfo.state = svcBuildStateHasOnlyRunningOrSucceededCBuilds
		return stateInfo, nil
	}

	stateInfo.state = svcBuildStateAllCBuildsSucceeded
	return stateInfo, nil
}
