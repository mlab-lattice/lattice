package systemlifecycle

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncAcceptedRollout(rollout *crv1.SystemRollout) error {
	build, err := c.systemBuildLister.SystemBuilds(rollout.Namespace).Get(rollout.Spec.BuildName)
	if err != nil {
		return err
	}

	switch build.Status.State {
	case crv1.SystemBuildStatePending, crv1.SystemBuildStateRunning:
		return nil

	case crv1.SystemBuildStateFailed:
		status := crv1.SystemRolloutStatus{
			State:   crv1.SystemRolloutStateFailed,
			Message: fmt.Sprintf("SystemBuild %v failed", build.Name),
		}

		// Copy rollout so the shared cache is not mutated
		rollout = rollout.DeepCopy()
		rollout.Status = status

		_, err := c.latticeClient.LatticeV1().SystemRollouts(rollout.Namespace).Update(rollout)
		if err != nil {
			return err
		}

		return c.relinquishRolloutOwningActionClaim(rollout)

	case crv1.SystemBuildStateSucceeded:
		system, err := c.getSystem(rollout.Namespace)
		if err != nil {
			return err
		}

		// Generate a fresh new System Spec
		spec, err := c.systemSpec(rollout, build)
		if err != nil {
			return err
		}

		// For each of the Services in the new System Spec, see if a Service already exists
		for path, svcInfo := range spec.Services {
			// If a Service already exists, use it.
			if existingSvcInfo, ok := system.Spec.Services[path]; ok {
				svcInfo.Name = existingSvcInfo.Name
				svcInfo.Status = existingSvcInfo.Status
				spec.Services[path] = svcInfo
			}
		}

		// Copy so the shared cache isn't mutated
		system = system.DeepCopy()
		system.Spec = *spec
		system.Status.State = crv1.SystemStateUpdating

		_, err = c.latticeClient.LatticeV1().Systems(system.Namespace).Update(system)
		if err != nil {
			return err
		}

		status := crv1.SystemRolloutStatus{
			State: crv1.SystemRolloutStateInProgress,
		}

		// Copy so the shared cache isn't mutated
		rollout = rollout.DeepCopy()
		rollout.Status = status

		_, err = c.latticeClient.LatticeV1().SystemRollouts(rollout.Namespace).Update(rollout)
		return err

	default:
		return fmt.Errorf("SystemBuild %v/%v in unexpected state %v", build.Namespace, build.Name, build.Status.State)
	}

	return nil
}
