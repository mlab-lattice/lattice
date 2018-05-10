package build

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) syncDeletedBuild(build *latticev1.Build) error {
	serviceBuilds, err := c.serviceBuildLister.ServiceBuilds(build.Namespace).List(labels.Everything())
	if err != nil {
		return fmt.Errorf("error getting service builds for deletion of %v: %v", build.Description(c.namespacePrefix), err)
	}

	for _, serviceBuild := range serviceBuilds {
		_, err := c.removeOwnerReference(build, serviceBuild)
		if err != nil {
			return err
		}
	}

	_, err = c.removeFinalizer(build)
	return err
}
