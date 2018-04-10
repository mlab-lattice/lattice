package system

import (
	"fmt"
	"reflect"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func (c *Controller) syncSystemStatus(
	system *latticev1.System,
	services map[tree.NodePath]latticev1.SystemStatusService,
	deletedServices []string,
) error {
	hasFailedService := false
	hasUpdatingService := false
	hasScalingService := false

	for path, status := range services {
		if status.State == latticev1.ServiceStateFailed {
			hasFailedService = true
			continue
		}

		if status.State == latticev1.ServiceStateUpdating || status.State == latticev1.ServiceStatePending {
			hasUpdatingService = true
			continue
		}

		if status.State == latticev1.ServiceStateScalingDown || status.State == latticev1.ServiceStateScalingUp {
			hasScalingService = true
			continue
		}

		if status.State != latticev1.ServiceStateStable {
			return fmt.Errorf("service %v (%v) had unexpected state: %v", path.ToDomain(), system.Name, status.State)
		}
	}

	state := latticev1.SystemStateStable

	// A scaling status takes priority over a stable status
	if hasScalingService {
		state = latticev1.SystemStateScaling
	}

	// An updating status takes priority over a scaling status
	if hasUpdatingService || len(deletedServices) != 0 {
		state = latticev1.SystemStateUpdating
	}

	// A failed status takes priority over an updating status
	if hasFailedService {
		state = latticev1.SystemStateDegraded
	}

	_, err := c.updateSystemStatus(system, state, services)
	return err
}

func (c *Controller) updateSystemStatus(
	system *latticev1.System,
	state latticev1.SystemState,
	services map[tree.NodePath]latticev1.SystemStatusService,
) (*latticev1.System, error) {
	status := latticev1.SystemStatus{
		State:              state,
		ObservedGeneration: system.Generation,
		// FIXME: remove this when ObservedGeneration is supported for CRD
		UpdateProcessed: true,
		Services:        services,
	}

	if reflect.DeepEqual(system.Status, status) {
		return system, nil
	}

	// Copy so the shared cache isn't mutated
	system = system.DeepCopy()
	system.Status = status

	return c.latticeClient.LatticeV1().Systems(system.Namespace).Update(system)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().Systems(system.Namespace).UpdateStatus(system)
}

func (c *Controller) removeFinalizer(system *latticev1.System) (*latticev1.System, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range system.Finalizers {
		if finalizer == constants.KubeFinalizerSystemController {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return system, nil
	}

	// The finalizer was in the list, so we should remove it.
	system = system.DeepCopy()
	system.Finalizers = finalizers

	return c.latticeClient.LatticeV1().Systems(system.Namespace).Update(system)
}
