package systemlifecycle

import (
	"fmt"
	"reflect"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) updateDeployStatus(
	deploy *latticev1.Deploy,
	state latticev1.DeployState,
	message string,
	internalError *string,
	buildID *v1.BuildID,
	path *tree.Path,
	version *v1.Version,
	startTimestamp *metav1.Time,
	completionTimestamp *metav1.Time,
) (*latticev1.Deploy, error) {
	status := latticev1.DeployStatus{
		State: state,

		Message:       message,
		InternalError: internalError,

		Build:   buildID,
		Path:    path,
		Version: version,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,
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

	// TODO(kevindrosendahl): switch to using actual system ID once they're UUIDs
	uniqueSystemIdentifier := v1.SystemID(namespace.UID)
	return c.lifecycleActions.AcquireDeploy(uniqueSystemIdentifier, deploy.V1ID(), path)
}

func (c *Controller) releaseDeployLock(deploy *latticev1.Deploy) error {
	namespace, err := c.kubeNamespaceLister.Get(deploy.Namespace)
	if err != nil {
		return err
	}

	// TODO(kevindrosendahl): switch to using actual system ID once they're UUIDs
	uniqueSystemIdentifier := v1.SystemID(namespace.UID)
	c.lifecycleActions.ReleaseDeploy(uniqueSystemIdentifier, deploy.V1ID())
	return nil
}
