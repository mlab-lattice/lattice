package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncPendingTeardown(teardown *latticev1.Teardown) error {
	currentOwningAction, err := c.attemptToClaimTeardownOwningAction(teardown)
	if err != nil {
		return err
	}

	if currentOwningAction != nil {
		_, err := c.updateTeardownStatus(
			teardown,
			latticev1.TeardownStateFailed,
			fmt.Sprintf("another lifecycle action is active: %v", currentOwningAction.String()),
		)
		return err
	}

	_, err = c.updateTeardownStatus(teardown, latticev1.TeardownStateInProgress, "")
	return err
}
