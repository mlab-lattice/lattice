package systemlifecycle

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncPendingRollout(rollout *crv1.SystemRollout) error {
	previousOwningAction := c.attemptToClaimRolloutOwningAction(rollout)

	status := crv1.SystemRolloutStatus{
		State: crv1.SystemRolloutStateAccepted,
	}
	if previousOwningAction != nil {
		status = crv1.SystemRolloutStatus{
			State:   crv1.SystemRolloutStateFailed,
			Message: fmt.Sprintf("another lifecycle action is active: %v", previousOwningAction.String()),
		}
	}

	// Copy so the shared cache isn't mutated
	rollout = rollout.DeepCopy()
	rollout.Status = status

	_, err := c.latticeClient.LatticeV1().SystemRollouts(rollout.Namespace).Update(rollout)
	return err
}
