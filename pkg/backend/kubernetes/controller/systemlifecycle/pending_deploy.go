package systemlifecycle

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
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
	now := metav1.Now()
	err := reflectutil.ValidateUnion(&deploy.Spec)
	if err != nil {
		switch err.(type) {
		case *reflectutil.InvalidUnionNoFieldSetError:
			_, err := c.updateDeployStatus(
				deploy,
				latticev1.DeployStateFailed,
				fmt.Sprintf("%v must specify build, path, or version", deploy.Description(c.namespacePrefix)),
				nil,
				nil,
				nil,
				nil,
				&now,
				&now,
			)
			return err

		case *reflectutil.InvalidUnionMultipleFieldSetError:
			_, err := c.updateDeployStatus(
				deploy,
				latticev1.DeployStateFailed,
				fmt.Sprintf("%v must only specify one of build, path, or version", deploy.Description(c.namespacePrefix)),
				nil,
				nil,
				nil,
				nil,
				&now,
				&now,
			)
			return err

		default:
			msg := err.Error()
			_, err := c.updateDeployStatus(
				deploy,
				latticev1.DeployStateFailed,
				"internal error",
				&msg,
				nil,
				nil,
				nil,
				&now,
				&now,
			)
			return err
		}
	}

	// get the deploy's path so we can attempt to acquire the proper lifecycle lock
	var buildID *v1.BuildID
	var path tree.Path
	var version *v1.Version
	switch {
	case deploy.Spec.Build != nil:
		buildID = deploy.Spec.Build
		path = tree.RootPath()

		build, err := c.buildLister.Builds(deploy.Namespace).Get(string(*buildID))
		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}

			build, err = c.latticeClient.LatticeV1().Builds(deploy.Namespace).Get(string(*buildID), metav1.GetOptions{})
			if err != nil {
				_, err := c.updateDeployStatus(
					deploy,
					latticev1.DeployStateFailed,
					fmt.Sprintf("%v build %v does not exist", deploy.Description(c.namespacePrefix), *buildID),
					nil,
					buildID,
					nil,
					nil,
					&now,
					&now,
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
				buildID,
				build.Spec.Path,
				build.Status.Version,
				&now,
				&now,
			)
			return err
		}

	case deploy.Spec.Path != nil:
		path = *deploy.Spec.Path

	case deploy.Spec.Version != nil:
		path = tree.RootPath()
		version = deploy.Spec.Version
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
			buildID,
			&path,
			version,
			&now,
			&now,
		)
		return err
	}

	_, err = c.updateDeployStatus(
		deploy,
		latticev1.DeployStateAccepted,
		"",
		nil,
		buildID,
		&path,
		version,
		&now,
		nil,
	)
	return err
}
