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

	activeCbs  []string
	failedCbs  []string
	needsNewCb []string
}

func (sbc *ServiceBuildController) calculateState(svcb *crv1.ServiceBuild) (*svcBuildStateInfo, error) {
	activeCbs := []string{}
	failedCbs := []string{}
	needsNewCbs := []string{}

	for component, cbInfo := range svcb.Spec.ComponentBuildsInfo {
		cb, exists, err := sbc.getComponentBuildFromInfo(&cbInfo, svcb.Namespace)
		if err != nil {
			return nil, err
		}

		if !exists {
			needsNewCbs = append(needsNewCbs, component)
			continue
		}

		switch cb.Status.State {
		case crv1.ComponentBuildStatePending, crv1.ComponentBuildStateQueued, crv1.ComponentBuildStateRunning:
			activeCbs = append(activeCbs, component)
		case crv1.ComponentBuildStateFailed:
			failedCbs = append(failedCbs, component)
		case crv1.ComponentBuildStateSucceeded:
			continue
		default:
			// FIXME: send warn event
			return nil, fmt.Errorf("ComponentBuild %v has unrecognized state %v", cb.Name, cb.Status.State)
		}
	}

	stateInfo := &svcBuildStateInfo{
		activeCbs:  activeCbs,
		failedCbs:  failedCbs,
		needsNewCb: needsNewCbs,
	}

	if len(failedCbs) > 0 {
		stateInfo.state = svcBuildStateHasFailedCBuilds
		return stateInfo, nil
	}

	if len(needsNewCbs) > 0 {
		stateInfo.state = svcBuildStateNoFailuresNeedsNewCBuilds
		return stateInfo, nil
	}

	if len(activeCbs) > 0 {
		stateInfo.state = svcBuildStateHasOnlyRunningOrSucceededCBuilds
		return stateInfo, nil
	}

	stateInfo.state = svcBuildStateAllCBuildsSucceeded
	return stateInfo, nil
}
