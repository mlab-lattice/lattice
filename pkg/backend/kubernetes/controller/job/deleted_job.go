package job

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncDeletedJob(job *latticev1.Job) error {
	_, err := c.removeFinalizer(job)
	return err
}
