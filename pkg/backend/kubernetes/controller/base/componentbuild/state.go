package componentbuild

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	batchv1 "k8s.io/api/batch/v1"
)

type state string

const (
	stateJobNotCreated state = "job-not-created"
	stateJobRunning    state = "job-running"
	stateJobSucceeded  state = "job-succeeded"
	stateJobFailed     state = "job-failed"
)

type stateInfo struct {
	state state
	job   *batchv1.Job
}

func (c *Controller) calculateState(build *crv1.ComponentBuild) (*stateInfo, error) {
	job, err := c.getJobForBuild(build)
	if err != nil {
		return nil, err
	}

	// FIXME: if a ComponentBuild was successful, but then for some reason the Job is deleted, should it still be
	// considered successful or should a new Job be spun up? Right now a new Job will be spun up.
	if job == nil {
		stateInfo := &stateInfo{
			state: stateJobNotCreated,
		}
		return stateInfo, nil
	}

	stateInfo := &stateInfo{
		job: job,
	}

	finished, succeeded := jobStatus(job)
	if !finished {
		stateInfo.state = stateJobRunning
		return stateInfo, nil
	}

	stateInfo.state = stateJobSucceeded
	if !succeeded {
		stateInfo.state = stateJobFailed
	}

	return stateInfo, nil
}
