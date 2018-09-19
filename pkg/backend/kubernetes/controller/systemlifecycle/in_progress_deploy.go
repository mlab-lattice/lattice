package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncInProgressDeploy(deploy *latticev1.Deploy) error {
	system, err := c.getSystem(deploy.Namespace)
	if err != nil {
		return err
	}

	// Check to see if the system controller has processed updates to its Spec.
	// If it hasn't, the system.Status.State is not up to date. Return no error
	// and wait until the System has been updated to resync.
	if !system.UpdateProcessed() {
		return nil
	}

	var state latticev1.DeployState
	switch system.Status.State {
	case latticev1.SystemStateUpdating, latticev1.SystemStateScaling:
		// Still in progress, nothing more to do
		return nil

	case latticev1.SystemStateStable:
		state = latticev1.DeployStateSucceeded

	case latticev1.SystemStateDegraded:
		state = latticev1.DeployStateFailed

	default:
		return fmt.Errorf("%v in unexpected state %v", system.Description(), system.Status.State)
	}

	now := metav1.Now()
	completionTimestamp := &now

	// need to update the deploy's status before releasing the lock. if we released the lock
	// first it's possible that the deployment status update could fail, and another deploy
	// successfully acquires the lock. if the controller then restarted, it could see
	// conflicting locks when seeding the lifecycle actions
	deploy, err = c.updateDeployStatus(
		deploy,
		state,
		"",
		nil,
		deploy.Status.Build,
		deploy.Status.StartTimestamp,
		completionTimestamp,
	)
	if err != nil {
		return err
	}

	return c.releaseDeployLock(deploy)
}
