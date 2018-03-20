package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncPendingTeardown(teardown *latticev1.Teardown) error {
	currentOwningAction := c.attemptToClaimTeardownOwningAction(teardown)
	if currentOwningAction != nil {
		_, err := c.updateTeardownStatus(
			teardown,
			latticev1.TeardownStateFailed,
			fmt.Sprintf("another lifecycle action is active: %v", currentOwningAction.String()),
		)
		return err
	}

	_, err := c.updateTeardownStatus(teardown, latticev1.TeardownStateInProgress, "")
	return err
}
