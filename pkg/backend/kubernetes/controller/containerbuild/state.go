package containerbuild

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

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

func (c *Controller) calculateState(build *latticev1.ContainerBuild) (*stateInfo, error) {
	job, err := c.getJobForBuild(build)
	if err != nil {
		return nil, err
	}

	if build.Status.State == latticev1.ContainerBuildStateFailed {
		stateInfo := &stateInfo{
			state: stateJobFailed,
			job:   job,
		}
		return stateInfo, nil
	}

	if build.Status.State == latticev1.ContainerBuildStateSucceeded {
		stateInfo := &stateInfo{
			state: stateJobSucceeded,
			job:   job,
		}
		return stateInfo, nil
	}

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
