package systemlifecycle

import (
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
)

func (slc *Controller) syncPendingRolloutState(sysRollout *crv1.SystemRollout) error {
	previousOwningAction, err := slc.attemptToClaimOwningRollout(sysRollout)
	if err != nil {
		return err
	}

	if previousOwningAction == nil {
		return nil
	}

	return slc.failRolloutDueToExistingOwningAction(sysRollout, previousOwningAction)
}

func (slc *Controller) attemptToClaimOwningRollout(sysRollout *crv1.SystemRollout) (*lifecycleAction, error) {
	slc.owningLifecycleActionsLock.Lock()
	defer slc.owningLifecycleActionsLock.Unlock()

	// TODO: should we check to see if the owning action is the same rollout?
	owningAction, ok := slc.owningLifecycleActions[sysRollout.Spec.LatticeNamespace]
	if !ok {
		// No owning owningRollout currently, we can claim it.
		return nil, slc.claimOwningRollout(sysRollout)
	}

	return owningAction, nil
}

func (slc *Controller) claimOwningRollout(sysRollout *crv1.SystemRollout) error {
	slc.owningLifecycleActions[sysRollout.Spec.LatticeNamespace] = &lifecycleAction{
		rollout: sysRollout,
	}

	newStatus := crv1.SystemRolloutStatus{
		State: crv1.SystemRolloutStateAccepted,
	}
	result, err := slc.updateSystemRolloutStatus(sysRollout, newStatus)
	if err != nil {
		return err
	}

	return slc.syncAcceptedRollout(result)
}

func (slc *Controller) failRolloutDueToExistingOwningAction(sysr *crv1.SystemRollout, owningAction *lifecycleAction) error {
	message := "Another lifecycle action is active: "
	if owningAction.rollout != nil {
		message += "rollout " + owningAction.rollout.Name
	} else {
		message += "teardown " + owningAction.teardown.Name
	}

	newStatus := crv1.SystemRolloutStatus{
		State:   crv1.SystemRolloutStateFailed,
		Message: message,
	}

	_, err := slc.updateSystemRolloutStatus(sysr, newStatus)
	return err
}
