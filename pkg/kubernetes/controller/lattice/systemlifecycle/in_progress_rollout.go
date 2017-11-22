package systemlifecycle

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
)

func (slc *SystemLifecycleController) syncInProgressRollout(sysRollout *crv1.SystemRollout) error {
	system, err := slc.getSystemForRollout(sysRollout)
	if err != nil {
		return err
	}

	if system == nil {
		// FIXME: send warn event
		// TODO: this seems kind of against the controller pattern, should we just move the system to an "accepted" state
		// 		 and resync instead?
		return fmt.Errorf("SystemRollout %v in-progress with no System", sysRollout.Name)
	}

	sysRollout, err = slc.syncRolloutWithSystem(sysRollout, system)
	if err != nil {
		// FIXME: is it possible that the rollout is locked forever now?
		return err
	}

	if sysRollout.Status.State == crv1.SystemRolloutStateSucceeded || sysRollout.Status.State == crv1.SystemRolloutStateFailed {
		return slc.relinquishOwningRolloutClaim(sysRollout)
	}

	return nil
}

func (slc *SystemLifecycleController) getSystemForRollout(sysRollout *crv1.SystemRollout) (*crv1.System, error) {
	var system *crv1.System

	latticeNamespace := sysRollout.Spec.LatticeNamespace
	for _, sysObj := range slc.systemStore.List() {
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

func (slc *SystemLifecycleController) syncRolloutWithSystem(sysRollout *crv1.SystemRollout, sys *crv1.System) (*crv1.SystemRollout, error) {
	var newState crv1.SystemRolloutStatus
	switch sys.Status.State {
	case crv1.SystemStateRollingOut:
		return sysRollout, nil
	case crv1.SystemStateRolloutSucceeded:
		newState = crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStateSucceeded,
		}
	case crv1.SystemStateRolloutFailed:
		newState = crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStateFailed,
		}
	}

	return slc.updateSystemRolloutStatus(sysRollout, newState)
}

func (slc *SystemLifecycleController) relinquishOwningRolloutClaim(sysRollout *crv1.SystemRollout) error {
	slc.owningRolloutsLock.Lock()
	defer slc.owningRolloutsLock.Unlock()

	owningRollout, ok := slc.owningRollouts[sysRollout.Spec.LatticeNamespace]
	if !ok || owningRollout.Name != sysRollout.Name {
		return fmt.Errorf("unexpected owning rollout %s in namespace %s", owningRollout.Name, sysRollout.Spec.LatticeNamespace)
	}

	delete(slc.owningRollouts, sysRollout.Spec.LatticeNamespace)
	return nil
}
