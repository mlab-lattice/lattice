package job

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncDeletedJobRun(jobRun *latticev1.JobRun) error {
	_, err := c.removeFinalizer(jobRun)
	return err
}
