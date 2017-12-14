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

	return c.latticeClient.LatticeV1().SystemTeardowns(teardown.Namespace).UpdateStatus(teardown)
}
