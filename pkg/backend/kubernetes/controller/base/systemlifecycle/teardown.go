package systemlifecycle

import (
	"reflect"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) updateTeardownStatus(teardown *crv1.SystemTeardown, state crv1.SystemTeardownState, message string) (*crv1.SystemTeardown, error) {
	status := crv1.SystemTeardownStatus{
		State:              state,
		ObservedGeneration: teardown.Generation,
		Message:            message,
	}

	if reflect.DeepEqual(teardown.Status, status) {
		return teardown, nil
	}

	// Copy so the shared cache isn't mutated
	teardown = teardown.DeepCopy()
	teardown.Status = status

	return c.latticeClient.LatticeV1().SystemTeardowns(teardown.Namespace).Update(teardown)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().SystemTeardowns(teardown.Namespace).UpdateStatus(teardown)
}
