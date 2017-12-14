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

	// Copy so the shared cache isn't mutated
	rollout = rollout.DeepCopy()
	rollout.Status.State = state

	rollout, err = c.latticeClient.LatticeV1().SystemRollouts(rollout.Namespace).Update(rollout)
	if err != nil {
		// FIXME: is it possible that the rollout is locked forever now?
		return err
	}

	if rollout.Status.State == crv1.SystemRolloutStateSucceeded || rollout.Status.State == crv1.SystemRolloutStateFailed {
		return c.relinquishRolloutOwningActionClaim(rollout)
	}

	return nil
}
