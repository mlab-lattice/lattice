package systemlifecycle

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncPendingRollout(rollout *crv1.SystemRollout) error {
	currentOwningAction := c.attemptToClaimRolloutOwningAction(rollout)
	if currentOwningAction != nil {
		_, err := c.updateRolloutStatus(rollout, crv1.SystemRolloutStateFailed, fmt.Sprintf("another lifecycle action is active: %v", currentOwningAction.String()))
		return err
	}

	_, err := c.updateRolloutStatus(rollout, crv1.SystemRolloutStateAccepted, "")
	return err
}
