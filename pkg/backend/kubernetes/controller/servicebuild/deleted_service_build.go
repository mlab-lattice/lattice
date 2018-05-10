package servicebuild

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) syncDeletedServiceBuild(build *latticev1.ServiceBuild) error {
	componentBuilds, err := c.componentBuildLister.ComponentBuilds(build.Namespace).List(labels.Everything())
	if err != nil {
		return fmt.Errorf("error getting component builds for deletion of %v: %v", build.Description(c.namespacePrefix), err)
	}

	for _, componentBuild := range componentBuilds {
		_, err := c.removeOwnerReference(build, componentBuild)
		if err != nil {
			return err
		}
	}

	_, err = c.removeFinalizer(build)
	return err
}
