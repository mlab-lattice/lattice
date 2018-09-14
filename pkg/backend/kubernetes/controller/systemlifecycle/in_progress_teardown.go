package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
)

func (c *Controller) syncInProgressTeardown(teardown *latticev1.Teardown) error {
	system, err := c.getSystem(teardown.Namespace)
	if err != nil {
		return err
	}

	definition := resolver.NewResolutionTree()
	artifacts := latticev1.NewSystemSpecWorkloadBuildArtifacts()

	system, err = c.updateSystem(system, definition, artifacts)
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

	// need to update the teardown's status before releasing the lock. if we released the lock
	// first it's possible that the teardown status update could fail, and another deploy or teardown
	// successfully acquires the lock. if the controller then restarted, it could see
	// conflicting locks when seeding the lifecycle actions
	teardown, err = c.updateTeardownStatus(teardown, state, "")
	if err != nil {
		return err
	}

	return c.releaseTeardownLock(teardown)
}
