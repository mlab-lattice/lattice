package systemlifecycle

import (
	"reflect"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) updateDeployStatus(
	deploy *latticev1.Deploy,
	state latticev1.DeployState,
	message string,
) (*latticev1.Deploy, error) {
	status := latticev1.DeployStatus{
		State:              state,
		ObservedGeneration: deploy.Generation,
		Message:            message,
	}

	if reflect.DeepEqual(deploy.Status, status) {
		return deploy, nil
	}

	// Copy so the shared cache isn't mutated
	deploy = deploy.DeepCopy()
	deploy.Status = status

	return c.latticeClient.LatticeV1().Deploies(deploy.Namespace).Update(deploy)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().SystemRollouts(deploy.Namespace).UpdateStatus(deploy)
}
