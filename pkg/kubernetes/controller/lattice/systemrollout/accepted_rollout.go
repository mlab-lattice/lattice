package systemrollout

import (
	"fmt"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
)

func (src *SystemRolloutController) syncAcceptedRollout(sysRollout *crv1.SystemRollout) error {
	sysBuild, err := src.getSystemBuildForRollout(sysRollout)
	if err != nil {
		return err
	}

	switch sysBuild.Status.State {
	case crv1.SystemBuildStateFailed:
		newStatus := crv1.SystemRolloutStatus{
			State:   crv1.SystemRolloutStateFailed,
			Message: fmt.Sprintf("SystemBuild %v failed", sysBuild.Name),
		}
		_, err := src.updateSystemRolloutStatus(sysRollout, newStatus)
		return err

	case crv1.SystemBuildStateSucceeded:
		sys, err := src.getSystemForRollout(sysRollout)
		if err != nil {
			return err
		}

		if sys == nil {
			sys, err = src.createSystem(sysRollout, sysBuild)
			if err != nil {
				return err
			}
		} else {
			// Generate a fresh new System Spec
			sysSpec, err := src.getNewSystemSpec(sysRollout, sysBuild)
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

			_, err = src.updateSystemSpec(sys, sysSpec)
			if err != nil {
				return err
			}
		}

		newStatus := crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStateInProgress,
		}

		result, err := src.updateSystemRolloutStatus(sysRollout, newStatus)
		if err != nil {
			return err
		}

		return src.syncInProgressRollout(result)
	}

	return nil
}

func (src *SystemRolloutController) getSystemBuildForRollout(sysRollout *crv1.SystemRollout) (*crv1.SystemBuild, error) {
	sysBuildKey := sysRollout.Namespace + "/" + sysRollout.Spec.BuildName
	sysBuildObj, exists, err := src.systemBuildStore.GetByKey(sysBuildKey)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("SystemBuild %v does not exist", sysBuildKey)
	}

	return sysBuildObj.(*crv1.SystemBuild), nil
}
