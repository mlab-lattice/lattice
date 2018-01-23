package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncAcceptedRollout(rollout *latticev1.SystemRollout) error {
	build, err := c.systemBuildLister.SystemBuilds(rollout.Namespace).Get(rollout.Spec.BuildName)
	if err != nil {
		return err
	}

	// Check to see if the system build controller has processed updates to its Spec.
	// If it hasn't, the build.Status.State is not up to date. Return no error
	// and wait until the System has been updated to resync.
	// TODO: don't think we actually need this here
	if !isSystemBuildStatusCurrent(build) {
		return nil
	}

	switch build.Status.State {
	case latticev1.SystemBuildStatePending, latticev1.SystemBuildStateRunning:
		return nil

	case latticev1.SystemBuildStateFailed:
		_, err := c.updateRolloutStatus(rollout, latticev1.SystemRolloutStateFailed, fmt.Sprintf("SystemBuild %v failed", build.Name))
		if err != nil {
			return err
		}

		return c.relinquishRolloutOwningActionClaim(rollout)

	case latticev1.SystemBuildStateSucceeded:
		system, err := c.getSystem(rollout.Namespace)
		if err != nil {
			return err
		}

		services, err := c.systemServices(rollout, build)
		if err != nil {
			return err
		}

		_, err = c.updateSystem(system, services)
		if err != nil {
			return err
		}

		_, err = c.updateRolloutStatus(rollout, latticev1.SystemRolloutStateInProgress, "")
		return err

	default:
		return fmt.Errorf("SystemBuild %v/%v in unexpected state %v", build.Namespace, build.Name, build.Status.State)
	}
}
