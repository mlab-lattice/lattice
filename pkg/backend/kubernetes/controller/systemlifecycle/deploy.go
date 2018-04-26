package systemlifecycle

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) updateDeployStatus(
	deploy *latticev1.Deploy,
	state latticev1.DeployState,
	message string,
) (*latticev1.Deploy, error) {
	status := latticev1.DeployStatus{
		State:   state,
		Message: message,
	}

	if reflect.DeepEqual(deploy.Status, status) {
		return deploy, nil
	}

	// Copy so the shared cache isn't mutated
	deploy = deploy.DeepCopy()
	deploy.Status = status

	result, err := c.latticeClient.LatticeV1().Deploys(deploy.Namespace).UpdateStatus(deploy)
	if err != nil {
		return nil, fmt.Errorf("error updating %v status: %v", deploy.Description(c.namespacePrefix), err)
	}

	return result, nil
}
