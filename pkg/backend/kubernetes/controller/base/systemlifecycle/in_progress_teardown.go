package systemlifecycle

import (
	"fmt"
	"reflect"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

func (c *Controller) syncInProgressTeardown(teardown *crv1.SystemTeardown) error {
	system, err := c.getSystem(teardown.Namespace)
	if err != nil {
		return err
	}

	spec := crv1.SystemSpec{
		Services: map[tree.NodePath]crv1.SystemSpecServiceInfo{},
	}

	// This needs to happen in here because we don't have an "Accepted" intermediate state like SystemRollout does.
	// We can't atomically update both teardown.Status.State and change system.Spec, and we need to move teardown into
	// "in progress" so on controller restart the controller can figure out that it is the owning object. Therefore,
	// we must update the teardown.Status.State first. If we were then to try to update system.Spec in syncPendingTeardown
	// and it failed, it would never get rerun since syncInProgressTeardown would always be called from there on out.
	// So instead we set the system.Spec in here to make sure it gets run even after failures.
	if !reflect.DeepEqual(system.Spec, spec) {
		// Copy so the shared cache isn't mutated
		system = system.DeepCopy()
		system.Spec = spec
		system.Status.State = crv1.SystemStateUpdating

		system, err = c.latticeClient.LatticeV1().Systems(system.Namespace).Update(system)
		if err != nil {
			return err
		}
	}

	var state crv1.SystemTeardownState
	switch system.Status.State {
	case crv1.SystemStateUpdating, crv1.SystemStateScaling:
		// Still in progress, nothing more to do
		return nil

	case crv1.SystemStateStable:
		state = crv1.SystemTeardownStateSucceeded

	case crv1.SystemStateFailed:
		state = crv1.SystemTeardownStateFailed

	default:
		return fmt.Errorf("System %v/%v in unexpected state %v", system.Namespace, system.Name, system.Status.State)
	}

	// Copy the teardown so the shared cache isn't mutated
	teardown = teardown.DeepCopy()
	teardown.Status.State = state

	teardown, err = c.latticeClient.LatticeV1().SystemTeardowns(teardown.Namespace).Update(teardown)
	if err != nil {
		// FIXME: is it possible that the teardown is locked forever now?
		return err
	}

	if teardown.Status.State == crv1.SystemTeardownStateSucceeded || teardown.Status.State == crv1.SystemTeardownStateFailed {
		return c.relinquishTeardownOwningActionClaim(teardown)
	}
	return nil
}
