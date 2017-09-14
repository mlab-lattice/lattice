package systemrollout

import (
	"fmt"
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

func (src *SystemRolloutController) syncAcceptedRollout(sysRollout *crv1.SystemRollout) error {
	systemBuild, err := src.getSystemBuildForRollout(sysRollout)
	if err != nil {
		return err
	}

	switch systemBuild.Status.State {
	case crv1.SystemBuildStateFailed:
		newStatus := crv1.SystemRolloutStatus{
			State:   crv1.SystemRolloutStateFailed,
			Message: fmt.Sprintf("SystemBuild %v failed", systemBuild.Name),
		}
		return src.updateStatus(sysRollout, newStatus)

	case crv1.SystemBuildStateSucceeded:
		newStatus := crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStateInProgress,
		}
		if err := src.updateStatus(sysRollout, newStatus); err != nil {
			return err
		}

		return src.syncInProgressRollout(sysRollout)
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
