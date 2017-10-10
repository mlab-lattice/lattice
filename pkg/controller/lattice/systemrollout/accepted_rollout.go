package systemrollout

import (
	"fmt"
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
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
		system, err := src.getSystemForRollout(sysRollout)
		if err != nil {
			return err
		}

		if system == nil {
			system, err = src.createSystem(sysRollout, sysBuild)
			if err != nil {
				return err
			}
		} else {
			sysSpec, err := src.getNewSystemSpec(sysRollout, sysBuild)
			if err != nil {
				return err
			}

			_, err = src.updateSystemSpec(system, sysSpec)
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
