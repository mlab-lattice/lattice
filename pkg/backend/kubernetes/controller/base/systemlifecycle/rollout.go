package systemlifecycle

import (
	"reflect"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) updateRolloutStatus(rollout *crv1.SystemRollout, state crv1.SystemRolloutState, message string) (*crv1.SystemRollout, error) {
	status := crv1.SystemRolloutStatus{
		State:              state,
		ObservedGeneration: rollout.Generation,
		Message:            message,
	}

	if reflect.DeepEqual(rollout.Status, status) {
		return rollout, nil
	}

	return c.latticeClient.LatticeV1().SystemRollouts(rollout.Namespace).UpdateStatus(rollout)
}
