package controller

import (
	"log"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func (c *Controller) runJob(job *v1.Job) {
	// add a little artificial delay before starting
	time.Sleep(time.Second)

	log.Printf("running job %v", job.ID)

	// change state to running
	func() {
		c.registry.Lock()
		defer c.registry.Unlock()
		now := time.Now()
		job.Status.State = v1.JobStateRunning
		job.Status.StartTimestamp = &now
	}()

	// sleep
	time.Sleep(7 * time.Second)

	c.registry.Lock()
	defer c.registry.Unlock()
	now := time.Now()
	job.Status.State = v1.JobStateSucceeded
	job.Status.CompletionTimestamp = &now
}
