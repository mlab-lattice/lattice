package systemlifecycle

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncPendingTeardown(teardown *crv1.SystemTeardown) error {
	currentOwningAction := c.attemptToClaimTeardownOwningAction(teardown)
	if currentOwningAction != nil {
		_, err := c.updateTeardownStatus(teardown, crv1.SystemTeardownStateFailed, fmt.Sprintf("another lifecycle action is active: %v", currentOwningAction.String()))
		return err
	}

	_, err := c.updateTeardownStatus(teardown, crv1.SystemTeardownStateInProgress, "")
	return err
}
