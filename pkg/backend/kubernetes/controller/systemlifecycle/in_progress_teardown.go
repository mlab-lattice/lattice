package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func (c *Controller) syncInProgressTeardown(teardown *latticev1.Teardown) error {
	system, err := c.getSystem(teardown.Namespace)
	if err != nil {
		return err
	}

	// This needs to happen in here because we don't have an "Accepted" intermediate state like Deploy does.
	// We can't atomically update both teardown.Status.State and change system.Spec, and we need to move teardown into
	// "in progress" so on controller restart the controller can figure out that it is the owning object. Therefore,
	// we must update the teardown.Status.State first. If we were then to try to update system.Spec in syncPendingTeardown
	// and it failed, it would never get rerun since syncInProgressTeardown would always be called from there on out.
	// So instead we set the system.Spec in here to make sure it gets run even after failures.
	services := map[tree.NodePath]latticev1.SystemSpecServiceInfo{}
	nodePools := map[string]latticev1.NodePoolSpec{}

	system, err = c.updateSystem(system, services, nodePools)
	if err != nil {
		return err
	}

	// Check to see if the system controller has processed updates to its Spec.
	// If it hasn't, the system.Status.State is not up to date. Return no error
	// and wait until the System has been updated to resync.
	if !system.UpdateProcessed() {
		return nil
	}

	var state latticev1.TeardownState
	switch system.Status.State {
	case latticev1.SystemStateUpdating, latticev1.SystemStateScaling:
		// Still in progress, nothing more to do
		return nil

	case latticev1.SystemStateStable:
		system, err = c.updateSystemLabels(system, nil, nil, nil)
		if err != nil {
			return err
		}

		state = latticev1.TeardownStateSucceeded

	case latticev1.SystemStateDegraded:
		state = latticev1.TeardownStateFailed

	default:
		return fmt.Errorf("%v in unexpected state %v", system.Description(), system.Status.State)
	}

	teardown, err = c.updateTeardownStatus(teardown, state, "")
	if err != nil {
		// FIXME: is it possible that the teardown is locked forever now?
		return err
	}

	if teardown.Status.State == latticev1.TeardownStateSucceeded || teardown.Status.State == latticev1.TeardownStateFailed {
		return c.relinquishTeardownOwningActionClaim(teardown)
	}
	return nil
}
