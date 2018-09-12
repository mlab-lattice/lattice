package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	reflectutil "github.com/mlab-lattice/lattice/pkg/util/reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncPendingDeploy(deploy *latticev1.Deploy) error {
	glog.V(5).Infof("syncing pending %v", deploy.Description(c.namespacePrefix))
	err := reflectutil.ValidateUnion(&deploy.Spec)
	if err != nil {
		switch err.(type) {
		case *reflectutil.InvalidUnionNoFieldSetError:
			_, err := c.updateDeployStatus(
				deploy,
				latticev1.DeployStateFailed,
				fmt.Sprintf("%v had must specify build, path, or version", deploy.Description(c.namespacePrefix)),
				nil,
				deploy.Status.BuildID,
			)
			return err

		case *reflectutil.InvalidUnionMultipleFieldSetError:
			_, err := c.updateDeployStatus(
				deploy,
				latticev1.DeployStateFailed,
				fmt.Sprintf("%v had must only specify one of build, path, or version", deploy.Description(c.namespacePrefix)),
				nil,
				deploy.Status.BuildID,
			)
			return err

		default:
			msg := err.Error()
			_, err := c.updateDeployStatus(
				deploy,
				latticev1.DeployStateFailed,
				"internal error",
				&msg,
				deploy.Status.BuildID,
			)
			return err
		}
	}

	// get the deploy's path so we can attempt to acquire the proper lifecycle lock
	var path tree.Path
	switch {
	case deploy.Spec.Build != nil:
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

		path = tree.RootPath()
		if build.Spec.Path != nil {
			path = *build.Spec.Path
		}

	case deploy.Spec.Path != nil:
		path = *deploy.Spec.Path

	case deploy.Spec.Version != nil:
		path = tree.RootPath()
	}

	// attempt to acquire the proper lifecycle lock for the deploy. if we fail due to a locking conflict,
	// fail the deploy.
	err = c.acquireDeployLock(deploy, path)
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
			nil,
		)
		return err
	}

	_, err = c.updateDeployStatus(deploy, latticev1.DeployStateAccepted, "", nil, deploy.Status.BuildID)
	return err
}
