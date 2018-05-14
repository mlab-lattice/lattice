package servicebuild

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncOrphanedServiceBuild(build *latticev1.ServiceBuild) error {
	return c.deleteServiceBuild(build)
}
