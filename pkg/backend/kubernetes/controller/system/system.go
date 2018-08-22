package system

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func (c *Controller) syncSystemStatus(
	system *latticev1.System,
	services map[tree.Path]latticev1.SystemStatusService,
	nodePools map[string]latticev1.SystemStatusNodePool,
) error {
	hasFailedService, hasUpdatingService, hasScalingService, err := servicesStateInfo(services)
	if err != nil {
		return fmt.Errorf("error getting services info for %v: %v", system.Description(), err)
	}

	hasFailedNodePool, hasUpdatingNodePool, hasScalingNodePool, err := nodePoolsStateInfo(nodePools)
	if err != nil {
		return fmt.Errorf("error getting node pools info for %v: %v", system.Description(), err)
	}

	state := latticev1.SystemStateStable

	// A scaling status takes priority over a stable status
	if hasScalingService || hasScalingNodePool {
		state = latticev1.SystemStateScaling
	}

	// An updating status takes priority over a scaling status
	if hasUpdatingService || hasUpdatingNodePool {
		state = latticev1.SystemStateUpdating
	}

	// A failed status takes priority over an updating status
	if hasFailedService || hasFailedNodePool {
		state = latticev1.SystemStateDegraded
	}

	_, err = c.updateSystemStatus(system, state, services)
	return err
}

func servicesStateInfo(services map[tree.Path]latticev1.SystemStatusService) (bool, bool, bool, error) {
	hasFailedService := false
	hasUpdatingService := false
	hasScalingService := false

	for path, status := range services {
		if status.ObservedGeneration < status.Generation {
			hasUpdatingService = true
			continue
		}

		switch status.State {
		case latticev1.ServiceStateFailed:
			hasFailedService = true

		case latticev1.ServiceStateScaling:
			hasScalingService = true

		case latticev1.ServiceStateUpdating, latticev1.ServiceStatePending, latticev1.ServiceStateDeleting:
			hasUpdatingService = true

		case latticev1.ServiceStateStable:
			// nothing to do

		default:
			return false, false, false, fmt.Errorf("service %v had unexpected state: %v", path.String(), status.State)
		}
	}

	return hasFailedService, hasUpdatingService, hasScalingService, nil
}

func nodePoolsStateInfo(nodePools map[string]latticev1.SystemStatusNodePool) (bool, bool, bool, error) {
	hasFailedNodePool := false
	hasUpdatingNodePool := false
	hasScalingNodePool := false

	for path, status := range nodePools {
		if status.ObservedGeneration < status.Generation {
			hasUpdatingNodePool = true
			continue
		}

		switch status.State {
		case latticev1.NodePoolStateFailed:
			hasFailedNodePool = true

		case latticev1.NodePoolStateScaling:
			hasScalingNodePool = true

		case latticev1.NodePoolStateUpdating, latticev1.NodePoolStatePending, latticev1.NodePoolStateDeleting:
			hasUpdatingNodePool = true

		case latticev1.NodePoolStateStable:
			// nothing to do

		default:
			return false, false, false, fmt.Errorf("node pool %v had unexpected state: %v", path, status.State)
		}
	}

	return hasFailedNodePool, hasUpdatingNodePool, hasScalingNodePool, nil
}

func (c *Controller) updateSystemStatus(
	system *latticev1.System,
	state latticev1.SystemState,
	services map[tree.Path]latticev1.SystemStatusService,
) (*latticev1.System, error) {
	status := latticev1.SystemStatus{
		ObservedGeneration: system.Generation,
		State:              state,
		Services:           services,
	}

	if reflect.DeepEqual(system.Status, status) {
		return system, nil
	}

	// Copy so the shared cache isn't mutated
	system = system.DeepCopy()
	system.Status = status

	result, err := c.latticeClient.LatticeV1().Systems(system.Namespace).UpdateStatus(system)
	if err != nil {
		return nil, fmt.Errorf("error updating %v status: %v", system.Description(), err)
	}

	return result, nil
}

func (c *Controller) addFinalizer(system *latticev1.System) (*latticev1.System, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range system.Finalizers {
		if finalizer == kubeutil.SystemControllerFinalizer {
			return system, nil
		}
	}

	// Copy so we don't mutate the shared cache
	system = system.DeepCopy()
	system.Finalizers = append(system.Finalizers, kubeutil.SystemControllerFinalizer)

	result, err := c.latticeClient.LatticeV1().Systems(system.Namespace).Update(system)
	if err != nil {
		return nil, fmt.Errorf("error adding %v finalizer: %v", system.Description(), err)
	}

	return result, nil
}

func (c *Controller) removeFinalizer(system *latticev1.System) (*latticev1.System, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range system.Finalizers {
		if finalizer == kubeutil.SystemControllerFinalizer {
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

	result, err := c.latticeClient.LatticeV1().Systems(system.Namespace).Update(system)
	if err != nil {
		return nil, fmt.Errorf("error removing finalizer from %v: %v", system.Description(), err)
	}

	return result, nil
}
