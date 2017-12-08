package systemlifecycle

import (
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
)

func (slc *Controller) syncPendingTeardown(syst *crv1.SystemTeardown) error {
	previousOwningAction, err := slc.attemptToClaimOwningTeardown(syst)
	if err != nil {
		return err
	}

	if previousOwningAction == nil {
		return nil
	}

	return slc.failTeardownDueToExistingOwningAction(syst, previousOwningAction)
}

func (slc *Controller) attemptToClaimOwningTeardown(syst *crv1.SystemTeardown) (*lifecycleAction, error) {
	slc.owningLifecycleActionsLock.Lock()
	defer slc.owningLifecycleActionsLock.Unlock()

	// TODO: should we check to see if the owning action is the same rollout?
	owningAction, ok := slc.owningLifecycleActions[syst.Spec.LatticeNamespace]
	if !ok {
		// No owning owningRollout currently, we can claim it.
		return nil, slc.claimOwningTeardown(syst)
	}

	return owningAction, nil
}

func (slc *Controller) claimOwningTeardown(syst *crv1.SystemTeardown) error {
	slc.owningLifecycleActions[syst.Spec.LatticeNamespace] = &lifecycleAction{
		teardown: syst,
	}

	newStatus := crv1.SystemTeardownStatus{
		State: crv1.SystemTeardownStateInProgress,
	}
	result, err := slc.updateSystemTeardownStatus(syst, newStatus)
	if err != nil {
		return err
	}

	return slc.syncInProgressTeardown(result)
}

func (slc *Controller) failTeardownDueToExistingOwningAction(syst *crv1.SystemTeardown, owningAction *lifecycleAction) error {
	message := "Another lifecycle action is active: "
	if owningAction.rollout != nil {
		message += "rollout " + owningAction.rollout.Name
	} else {
		message += "teardown " + owningAction.teardown.Name
	}

	newStatus := crv1.SystemTeardownStatus{
		State:   crv1.SystemTeardownStateFailed,
		Message: message,
	}

	_, err := slc.updateSystemTeardownStatus(syst, newStatus)
	return err
}
