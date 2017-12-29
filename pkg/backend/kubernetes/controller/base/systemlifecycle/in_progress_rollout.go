package systemlifecycle

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncInProgressRollout(rollout *crv1.SystemRollout) error {
	system, err := c.getSystem(rollout.Namespace)
	if err != nil {
		return err
	}

	// Check to see if the system controller has processed updates to its Spec.
	// If it hasn't, the system.Status.State is not up to date. Return no error
	// and wait until the System has been updated to resync.
	if !isSystemStatusCurrent(system) {
		return nil
	}

	var state crv1.SystemRolloutState
	switch system.Status.State {
	case crv1.SystemStateUpdating, crv1.SystemStateScaling:
		// Still in progress, nothing more to do
		return nil

	case crv1.SystemStateStable:
		state = crv1.SystemRolloutStateSucceeded

	case crv1.SystemStateFailed:
		state = crv1.SystemRolloutStateFailed

	default:
		return fmt.Errorf("System %v/%v in unexpected state %v", system.Namespace, system.Name, system.Status.State)
	}

	rollout, err = c.updateRolloutStatus(rollout, state, "")
	if err != nil {
		// FIXME: is it possible that the rollout is locked forever now?
		return err
	}

	if rollout.Status.State == crv1.SystemRolloutStateSucceeded || rollout.Status.State == crv1.SystemRolloutStateFailed {
		return c.relinquishRolloutOwningActionClaim(rollout)
	}

	return nil
}
