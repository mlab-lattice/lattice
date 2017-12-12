package systembuild

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"
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

	successfulSvcbs map[tree.NodePath]*crv1.ServiceBuild
	activeSvcbs     map[tree.NodePath]*crv1.ServiceBuild
	failedSvcbs     map[tree.NodePath]*crv1.ServiceBuild
	needsNewSvcb    []tree.NodePath
}

func (sbc *Controller) calculateState(sysb *crv1.SystemBuild) (*sysBuildStateInfo, error) {
	successfulSvcbs := map[tree.NodePath]*crv1.ServiceBuild{}
	activeSvcbs := map[tree.NodePath]*crv1.ServiceBuild{}
	failedSvcbs := map[tree.NodePath]*crv1.ServiceBuild{}
	needsNewSvcbs := []tree.NodePath{}

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
			activeSvcbs[service] = svcb
		case crv1.ServiceBuildStateFailed:
			failedSvcbs[service] = svcb
		case crv1.ServiceBuildStateSucceeded:
			successfulSvcbs[service] = svcb
		default:
			// FIXME: send warn event
			return nil, fmt.Errorf("SystemBuild %v has unrecognized state %v", svcb.Name, svcb.Status.State)
		}
	}

	stateInfo := &sysBuildStateInfo{
		successfulSvcbs: successfulSvcbs,
		activeSvcbs:     activeSvcbs,
		failedSvcbs:     failedSvcbs,
		needsNewSvcb:    needsNewSvcbs,
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
