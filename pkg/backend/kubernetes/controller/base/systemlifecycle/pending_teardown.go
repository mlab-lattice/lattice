package systemlifecycle

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncPendingTeardown(teardown *crv1.SystemTeardown) error {
	previousOwningAction := c.attemptToClaimTeardownOwningAction(teardown)

	status := crv1.SystemTeardownStatus{
		State: crv1.SystemTeardownStateInProgress,
	}
	if previousOwningAction != nil {
		status = crv1.SystemTeardownStatus{
			State:   crv1.SystemTeardownStateFailed,
			Message: fmt.Sprintf("another lifecycle action is active: %v", previousOwningAction.String()),
		}
	}

	// Copy so the shared cache isn't mutated
	teardown = teardown.DeepCopy()
	teardown.Status = status

	_, err := c.latticeClient.LatticeV1().SystemTeardowns(teardown.Namespace).Update(teardown)
	return err
}
