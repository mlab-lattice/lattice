package systemlifecycle

import (
	"fmt"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (slc *Controller) syncAcceptedRollout(sysRollout *crv1.SystemRollout) error {
	sysBuild, err := slc.getSystemBuildForRollout(sysRollout)
	if err != nil {
		return err
	}

	switch sysBuild.Status.State {
	case crv1.SystemBuildStateFailed:
		newStatus := crv1.SystemRolloutStatus{
			State:   crv1.SystemRolloutStateFailed,
			Message: fmt.Sprintf("SystemBuild %v failed", sysBuild.Name),
		}
		_, err := slc.updateSystemRolloutStatus(sysRollout, newStatus)
		if err != nil {
			return err
		}

		return slc.relinquishOwningRolloutClaim(sysRollout)

	case crv1.SystemBuildStateSucceeded:
		sys, err := slc.getSystemForRollout(sysRollout)
		if err != nil {
			return err
		}

		if sys == nil {
			sys, err = slc.createSystem(sysRollout, sysBuild)
			if err != nil {
				return err
			}
		} else {
			// Generate a fresh new System Spec
			sysSpec, err := slc.getNewSystemSpec(sysRollout, sysBuild)
			if err != nil {
				return err
			}

			// For each of the Services in the new System Spec, see if a Service already exists
			for path, svcInfo := range sysSpec.Services {
				// If a Service already exists, use it.
				if existingSvcInfo, ok := sys.Spec.Services[path]; ok {
					svcInfo.ServiceName = existingSvcInfo.ServiceName
					sysSpec.Services[path] = svcInfo
				}
			}

			_, err = slc.updateSystemSpec(sys, sysSpec)
			if err != nil {
				return err
			}
		}

		newStatus := crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStateInProgress,
		}

		result, err := slc.updateSystemRolloutStatus(sysRollout, newStatus)
		if err != nil {
			return err
		}

		return slc.syncInProgressRollout(result)
	}

	return nil
}

func (slc *Controller) getSystemBuildForRollout(sysRollout *crv1.SystemRollout) (*crv1.SystemBuild, error) {
	return slc.systemBuildLister.SystemBuilds(sysRollout.Namespace).Get(sysRollout.Spec.BuildName)
}
