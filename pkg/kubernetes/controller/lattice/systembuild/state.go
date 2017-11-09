package systembuild

import (
	"fmt"

	systemtree "github.com/mlab-lattice/core/pkg/system/tree"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
)

type sysBuildState string

const (
	sysBuildStateHasFailedCBuilds                 sysBuildState = "has-failed-svcbuilds"
	sysBuildStateHasOnlyRunningOrSucceededCBuilds sysBuildState = "has-only-succeeded-or-running-svcbuilds"
	sysBuildStateNoFailuresNeedsNewCBuilds        sysBuildState = "no-failures-needs-new-csvbuilds"
	sysBuildStateAllCBuildsSucceeded              sysBuildState = "all-svcbuilds-succeeded"
)

type sysBuildStateInfo struct {
	state sysBuildState

	activeSvcbs  []systemtree.NodePath
	failedSvcbs  []systemtree.NodePath
	needsNewSvcb []systemtree.NodePath
}

func (sbc *SystemBuildController) calculateState(sysb *crv1.SystemBuild) (*sysBuildStateInfo, error) {
	activeSvcbs := []systemtree.NodePath{}
	failedSvcbs := []systemtree.NodePath{}
	needsNewSvcbs := []systemtree.NodePath{}

	for service, svcbInfo := range sysb.Spec.Services {
		svcb, exists, err := sbc.getServiceBuildFromInfo(&svcbInfo, sysb.Namespace)
		if err != nil {
			return nil, err
		}

		if !exists {
			needsNewSvcbs = append(needsNewSvcbs, service)
			continue
		}

		switch svcb.Status.State {
		case crv1.ServiceBuildStatePending, crv1.ServiceBuildStateRunning:
			activeSvcbs = append(activeSvcbs, service)
		case crv1.ServiceBuildStateFailed:
			failedSvcbs = append(failedSvcbs, service)
		case crv1.ServiceBuildStateSucceeded:
			continue
		default:
			// FIXME: send warn event
			return nil, fmt.Errorf("SystemBuild %v has unrecognized state %v", svcb.Name, svcb.Status.State)
		}
	}

	stateInfo := &sysBuildStateInfo{
		activeSvcbs:  activeSvcbs,
		failedSvcbs:  failedSvcbs,
		needsNewSvcb: needsNewSvcbs,
	}

	if len(failedSvcbs) > 0 {
		stateInfo.state = sysBuildStateHasFailedCBuilds
		return stateInfo, nil
	}

	if len(needsNewSvcbs) > 0 {
		stateInfo.state = sysBuildStateNoFailuresNeedsNewCBuilds
		return stateInfo, nil
	}

	if len(activeSvcbs) > 0 {
		stateInfo.state = sysBuildStateHasOnlyRunningOrSucceededCBuilds
		return stateInfo, nil
	}

	stateInfo.state = sysBuildStateAllCBuildsSucceeded
	return stateInfo, nil
}
