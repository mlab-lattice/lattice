package systemlifecycle

import (
	"fmt"
	"reflect"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func (c *Controller) updateDeployStatus(
	deploy *latticev1.Deploy,
	state latticev1.DeployState,
	message string,
	buildID *v1.BuildID,
) (*latticev1.Deploy, error) {
	status := latticev1.DeployStatus{
		State:   state,
		Message: message,

		BuildID: buildID,
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

func (c *Controller) acquireDeployLock(deploy *latticev1.Deploy, path tree.Path) error {
	namespace, err := c.kubeNamespaceLister.Get(deploy.Namespace)
	if err != nil {
		return err
	}

	return c.lifecycleActions.AcquireDeploy(namespace.UID, deploy.V1ID(), deploy.Spec.Version.Path)
}

func (c *Controller) releaseDeployLock(deploy *latticev1.Deploy) error {
	namespace, err := c.kubeNamespaceLister.Get(deploy.Namespace)
	if err != nil {
		return err
	}

	c.lifecycleActions.ReleaseDeploy(namespace.UID, deploy.V1ID())
	return nil
}
