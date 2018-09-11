package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncPendingDeploy(deploy *latticev1.Deploy) error {
	glog.V(5).Infof("syncing pending %v", deploy.Description(c.namespacePrefix))
	if deploy.Spec.Version == nil && deploy.Spec.Build == nil {
		_, err := c.updateDeployStatus(
			deploy,
			latticev1.DeployStateFailed,
			fmt.Sprintf("%v had neither version nor build id", deploy.Description(c.namespacePrefix)),
			deploy.Status.BuildID,
		)
		return err
	}

	// get the deploy's path so we can attempt to acquire the proper lifecycle lock
	var path tree.Path
	if deploy.Spec.Version != nil {
		path = deploy.Spec.Version.Path
	} else {
		buildID := *deploy.Spec.Build
		build, err := c.buildLister.Builds(deploy.Namespace).Get(string(buildID))
		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}

			build, err = c.latticeClient.LatticeV1().Builds(deploy.Namespace).Get(string(buildID), metav1.GetOptions{})
			if err != nil {
				return err
			}
		}

		path = build.Spec.Path
	}

	// attempt to acquire the proper lifecycle lock for the deploy. if we fail due to a locking conflict,
	// fail the deploy.
	err := c.acquireDeployLock(deploy, path)
	if err != nil {
		_, ok := err.(*conflictingLifecycleActionError)
		if !ok {
			return err
		}

		_, err = c.updateDeployStatus(
			deploy,
			latticev1.DeployStateFailed,
			fmt.Sprintf("unable to acquire lifecycle lock: %v", err.Error()),
			nil,
		)
		return err
	}

	_, err = c.updateDeployStatus(deploy, latticev1.DeployStateAccepted, "", deploy.Status.BuildID)
	return err
}
