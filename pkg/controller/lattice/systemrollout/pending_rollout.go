package systemrollout

import (
	coretypes "github.com/mlab-lattice/core/pkg/types"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

func (src *SystemRolloutController) syncPendingRolloutState(sysRollout *crv1.SystemRollout) error {
	activeOwningRollout := src.checkIfOwningRolloutActive(sysRollout.Spec.LatticeNamespace)
	if !activeOwningRollout {
		return src.attemptToClaimOwningRollout(sysRollout)
	}

	return src.failRolloutDueToExistingActiveRollout(sysRollout)
}

func (src *SystemRolloutController) checkIfOwningRolloutActive(latticeNamespace coretypes.LatticeNamespace) bool {
	src.owningRolloutsLock.RLock()
	defer src.owningRolloutsLock.RUnlock()

	rollout, ok := src.owningRollouts[latticeNamespace]
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

func (src *SystemRolloutController) attemptToClaimOwningRollout(sysRollout *crv1.SystemRollout) error {
	src.owningRolloutsLock.Lock()
	defer src.owningRolloutsLock.Unlock()

	owningRollout, ok := src.owningRollouts[sysRollout.Spec.LatticeNamespace]
	if !ok {
		// No owning owningRollout currently, we can claim it.
		return src.claimOwningRollout(sysRollout)
	}

	// TODO: check to ensure that the SystemRollout obj pointed to in owningRollouts gets its state
	// updated when the actual SystemRollout's state is updated. We may want to do a read from the API here.
	rolloutState := owningRollout.Status.State
	if rolloutState == crv1.SystemRolloutStateSucceeded || rolloutState == crv1.SystemRolloutStateFailed {
		return src.claimOwningRollout(sysRollout)
	}

	return src.failRolloutDueToExistingActiveRollout(sysRollout)
}

func (src *SystemRolloutController) claimOwningRollout(sysRollout *crv1.SystemRollout) error {
	src.owningRollouts[sysRollout.Spec.LatticeNamespace] = sysRollout

	newStatus := crv1.SystemRolloutStatus{
		State: crv1.SystemRolloutStateAccepted,
	}
	result, err := src.updateSystemRolloutStatus(sysRollout, newStatus)
	if err != nil {
		return err
	}

	return src.syncAcceptedRollout(result)
}

func (src *SystemRolloutController) failRolloutDueToExistingActiveRollout(sysRollout *crv1.SystemRollout) error {
	newStatus := crv1.SystemRolloutStatus{
		State:   crv1.SystemRolloutStateFailed,
		Message: "Another SystemRollout is active",
	}

	_, err := src.updateSystemRolloutStatus(sysRollout, newStatus)
	return err
}
