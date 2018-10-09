package controller

import (
	"log"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/registry"
	timeutil "github.com/mlab-lattice/lattice/pkg/util/time"

	"github.com/satori/go.uuid"
)

func (c *Controller) runJob(job *v1.Job, record *registry.SystemRecord) {
	// add a little artificial delay before starting
	time.Sleep(time.Second)

	log.Printf("running job %v", job.ID)

	runID := v1.JobRunID(uuid.NewV4().String())

	// change state to running
	func() {
		c.registry.Lock()
		defer c.registry.Unlock()
		now := timeutil.New(time.Now())
		job.Status.State = v1.JobStateRunning
		job.Status.StartTimestamp = now

		run := v1.JobRun{
			ID: runID,

			Status: v1.JobRunStatus{
				State: v1.JobRunStateRunning,

				StartTimestamp: now,
			},
		}
		record.Jobs[job.ID].Runs[run.ID] = run
	}()

	// sleep
	time.Sleep(7 * time.Second)

	c.registry.Lock()
	defer c.registry.Unlock()
	now := timeutil.New(time.Now())
	job.Status.State = v1.JobStateSucceeded
	job.Status.CompletionTimestamp = now

	zero := int32(0)
	run := record.Jobs[job.ID].Runs[runID]
	run.Status.State = v1.JobRunStateSucceeded
	run.Status.CompletionTimestamp = now
	run.Status.ExitCode = &zero
}
