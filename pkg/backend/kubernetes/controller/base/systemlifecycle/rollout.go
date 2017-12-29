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

	// Copy so the shared cache isn't mutated
	rollout = rollout.DeepCopy()
	rollout.Status = status

	return c.latticeClient.LatticeV1().SystemRollouts(rollout.Namespace).Update(rollout)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().SystemRollouts(rollout.Namespace).UpdateStatus(rollout)
}
