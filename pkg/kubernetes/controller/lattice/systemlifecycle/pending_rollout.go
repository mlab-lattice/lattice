package systemlifecycle

import (
	coretypes "github.com/mlab-lattice/core/pkg/types"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
)

func (slc *SystemLifecycleController) syncPendingRolloutState(sysRollout *crv1.SystemRollout) error {
	activeOwningRollout := slc.checkIfOwningRolloutActive(sysRollout.Spec.LatticeNamespace)
	if !activeOwningRollout {
		return slc.attemptToClaimOwningRollout(sysRollout)
	}

	return slc.failRolloutDueToExistingActiveRollout(sysRollout)
}

func (slc *SystemLifecycleController) checkIfOwningRolloutActive(latticeNamespace coretypes.LatticeNamespace) bool {
	slc.owningRolloutsLock.RLock()
	defer slc.owningRolloutsLock.RUnlock()

	rollout, ok := slc.owningRollouts[latticeNamespace]
	if !ok {
		return false
	}

	// TODO: check to ensure that the SystemRollout obj pointed to in owningRollouts gets its state
	// updated when the actual SystemRollout's state is updated. We may want to do a read from the API here.
	rolloutState := rollout.Status.State
	if rolloutState == crv1.SystemRolloutStateSucceeded || rolloutState == crv1.SystemRolloutStateFailed {
		return false
	}

	return true
}

func (slc *SystemLifecycleController) attemptToClaimOwningRollout(sysRollout *crv1.SystemRollout) error {
	slc.owningRolloutsLock.Lock()
	defer slc.owningRolloutsLock.Unlock()

	owningRollout, ok := slc.owningRollouts[sysRollout.Spec.LatticeNamespace]
	if !ok {
		// No owning owningRollout currently, we can claim it.
		return slc.claimOwningRollout(sysRollout)
	}

	// TODO: check to ensure that the SystemRollout obj pointed to in owningRollouts gets its state
	// updated when the actual SystemRollout's state is updated. We may want to do a read from the API here.
	rolloutState := owningRollout.Status.State
	if rolloutState == crv1.SystemRolloutStateSucceeded || rolloutState == crv1.SystemRolloutStateFailed {
		return slc.claimOwningRollout(sysRollout)
	}

	return slc.failRolloutDueToExistingActiveRollout(sysRollout)
}

func (slc *SystemLifecycleController) claimOwningRollout(sysRollout *crv1.SystemRollout) error {
	slc.owningRollouts[sysRollout.Spec.LatticeNamespace] = sysRollout

	newStatus := crv1.SystemRolloutStatus{
		State: crv1.SystemRolloutStateAccepted,
	}
	result, err := slc.updateSystemRolloutStatus(sysRollout, newStatus)
	if err != nil {
		return err
	}

	return slc.syncAcceptedRollout(result)
}

func (slc *SystemLifecycleController) failRolloutDueToExistingActiveRollout(sysRollout *crv1.SystemRollout) error {
	newStatus := crv1.SystemRolloutStatus{
		State:   crv1.SystemRolloutStateFailed,
		Message: "Another SystemRollout is active",
	}

	_, err := slc.updateSystemRolloutStatus(sysRollout, newStatus)
	return err
}
