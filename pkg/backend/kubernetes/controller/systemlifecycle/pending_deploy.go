package systemlifecycle

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	reflectutil "github.com/mlab-lattice/lattice/pkg/util/reflect"
	syncutil "github.com/mlab-lattice/lattice/pkg/util/sync"

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
				fmt.Sprintf("%v must specify build, path, or version", deploy.Description(c.namespacePrefix)),
				nil,
				deploy.Status.Build,
				nil,
				nil,
			)
			return err

		case *reflectutil.InvalidUnionMultipleFieldSetError:
			_, err := c.updateDeployStatus(
				deploy,
				latticev1.DeployStateFailed,
				fmt.Sprintf("%v must only specify one of build, path, or version", deploy.Description(c.namespacePrefix)),
				nil,
				deploy.Status.Build,
				nil,
				nil,
			)
			return err

		default:
			msg := err.Error()
			_, err := c.updateDeployStatus(
				deploy,
				latticev1.DeployStateFailed,
				"internal error",
				&msg,
				deploy.Status.Build,
				nil,
				nil,
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
				_, err := c.updateDeployStatus(
					deploy,
					latticev1.DeployStateFailed,
					fmt.Sprintf("%v build %v does not exist", deploy.Description(c.namespacePrefix), buildID),
					nil,
					deploy.Status.Build,
					nil,
					nil,
				)
				return err
			}
		}

		// There are many factors that could make an old partial system build incompatible with the
		// current state of the system. Instead of trying to enumerate and handle them, for now
		// we'll simply fail the deploy.
		// May want to revisit this.
		if build.Spec.Path != nil {
			_, err := c.updateDeployStatus(
				deploy,
				latticev1.DeployStateFailed,
				fmt.Sprintf("cannot deploy using a build id (%v) since it is only a partial system build", buildID),
				nil,
				deploy.Status.Build,
				nil,
				nil,
			)
			return err
		}

		path = tree.RootPath()

	case deploy.Spec.Path != nil:
		path = *deploy.Spec.Path

	case deploy.Spec.Version != nil:
		path = tree.RootPath()
	}

	// attempt to acquire the proper lifecycle lock for the deploy. if we fail due to a locking conflict,
	// fail the deploy.
	err = c.acquireDeployLock(deploy, path)
	if err != nil {
		_, ok := err.(*syncutil.ConflictingLifecycleActionError)
		if !ok {
			return err
		}

		_, err = c.updateDeployStatus(
			deploy,
			latticev1.DeployStateFailed,
			fmt.Sprintf("unable to acquire lifecycle lock: %v", err.Error()),
			nil,
			nil,
			nil,
			nil,
		)
		return err
	}

	now := metav1.Now()
	startTimestamp := &now
	completionTimestamp := &now

	_, err = c.updateDeployStatus(
		deploy,
		latticev1.DeployStateAccepted,
		"",
		nil,
		deploy.Status.Build,
		startTimestamp,
		completionTimestamp,
	)
	return err
}
