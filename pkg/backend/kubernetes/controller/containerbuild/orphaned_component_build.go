package containerbuild

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncOrphanedComponentBuild(build *latticev1.ContainerBuild) error {
	return c.deleteComponentBuild(build)
}
