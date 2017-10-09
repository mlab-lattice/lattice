package systemrollout

import (
	"fmt"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

func (src *SystemRolloutController) syncInProgressRollout(sysRollout *crv1.SystemRollout) error {
	system, err := src.getSystemForRollout(sysRollout)
	if err != nil {
		return err
	}

	if system == nil {
		// FIXME: send warn event
		// TODO: this seems kind of against the controller pattern, should we just move the system to an "accepted" state
		// 		 and resync instead?
		return fmt.Errorf("SystemRollout %v in-progress with no System", sysRollout.Name)
	}

	return src.syncRolloutWithSystem(sysRollout, system)
}

func (src *SystemRolloutController) getSystemForRollout(sysRollout *crv1.SystemRollout) (*crv1.System, error) {
	var system *crv1.System

	latticeNamespace := sysRollout.Spec.LatticeNamespace
	for _, sysObj := range src.systemStore.List() {
		sys := sysObj.(*crv1.System)

		if string(latticeNamespace) == sys.Namespace {
			if system != nil {
				return nil, fmt.Errorf("LatticeNamespace %v contains multiple Systems", latticeNamespace)
			}

			system = sys
		}
	}

	return system, nil
}

func (src *SystemRolloutController) syncRolloutWithSystem(sysRollout *crv1.SystemRollout, sys *crv1.System) error {
	var newState crv1.SystemRolloutStatus
	switch sys.Status.State {
	case crv1.SystemStateRollingOut:
		return nil
	case crv1.SystemStateRolloutSucceeded:
		newState = crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStateSucceeded,
		}
	case crv1.SystemStateRolloutFailed:
		newState = crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStateFailed,
		}
	}

	_, err := src.updateSystemRolloutStatus(sysRollout, newState)
	return err
}
