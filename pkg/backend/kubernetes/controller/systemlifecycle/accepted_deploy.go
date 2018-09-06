package systemlifecycle

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (c *Controller) syncAcceptedDeploy(deploy *latticev1.Deploy) error {
	if deploy.Spec.Version != nil {
		build := &latticev1.Build{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: deploy.Namespace,
				Name:      uuid.NewV4().String(),
			},
			Spec: latticev1.BuildSpec{
				Version: deploy.Spec.Version.Version,
				Path:    deploy.Spec.Version.Path,
			},
		}

		_, err := c.latticeClient.LatticeV1().Builds(deploy.Namespace).Create(build)
		return err
	}

	if deploy.Spec.Build == nil {
		return fmt.Errorf("%v had neither version information or a build ID", deploy.Description(c.namespacePrefix))
	}

	build, err := c.buildLister.Builds(deploy.Namespace).Get(string(*deploy.Spec.Build))
	if err != nil {
		return fmt.Errorf(
			"error getting build %v for %v: %v",
			deploy.Spec.Build,
			deploy.Description(c.namespacePrefix),
			err,
		)
	}

	switch build.Status.State {
	case latticev1.BuildStatePending, latticev1.BuildStateAccepted, latticev1.BuildStateRunning:
		return nil

	case latticev1.BuildStateFailed:
		_, err := c.updateDeployStatus(
			deploy,
			latticev1.DeployStateFailed,
			fmt.Sprintf("%v failed", build.Description(c.namespacePrefix)),
		)
		if err != nil {
			return err
		}

		return c.relinquishDeployOwningActionClaim(deploy)

	case latticev1.BuildStateSucceeded:
		system, err := c.getSystem(deploy.Namespace)
		if err != nil {
			return err
		}

		version := v1.SystemVersion("unknown")
		if label, ok := deploy.DefinitionVersionLabel(); ok {
			version = label
		}

		buildID := v1.BuildID("unknown")
		if label, ok := deploy.BuildIDLabel(); ok {
			buildID = label
		}

		deployID := v1.DeployID(deploy.Name)

		system, err = c.updateSystemLabels(system, &version, &deployID, &buildID)
		if err != nil {
			return err
		}

		services, err := c.systemServices(build)
		if err != nil {
			return fmt.Errorf("error getting services for %v: %v", build.Description(c.namespacePrefix), err)
		}

		jobs, err := c.systemJobs(build)
		if err != nil {
			return fmt.Errorf("error getting jobs for %v: %v", build.Description(c.namespacePrefix), err)
		}

		nodePools, err := c.systemNodePools(build)
		if err != nil {
			return fmt.Errorf("error getting node pools for %v: %v", build.Description(c.namespacePrefix), err)
		}

		_, err = c.updateSystem(system, services, jobs, nodePools)
		if err != nil {
			return err
		}

		_, err = c.updateDeployStatus(deploy, latticev1.DeployStateInProgress, "")
		return err

	default:
		return fmt.Errorf("%v in unexpected state %v", build.Description(c.namespacePrefix), build.Status.State)
	}
}
