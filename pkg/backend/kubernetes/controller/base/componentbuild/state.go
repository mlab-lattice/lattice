package componentbuild

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	batchv1 "k8s.io/api/batch/v1"
)

type cBuildState string

const (
	stateJobNotCreated cBuildState = "job-not-created"
	stateJobRunning    cBuildState = "job-running"
	stateJobSucceeded  cBuildState = "job-succeeded"
	stateJobFailed     cBuildState = "job-failed"
)

type cBuildStateInfo struct {
	state cBuildState
	job   *batchv1.Job
}

func (c *Controller) calculateState(cb *crv1.ComponentBuild) (*cBuildStateInfo, error) {
	j, err := c.getJobForBuild(cb)
	if err != nil {
		return nil, err
	}

	// FIXME: if a ComponentBuild was successful, but then for some reason the Job is deleted, should it still be
	// considered successful or should a new Job be spun up? Right now a new Job will be spun up.
	if j == nil {
		stateInfo := &cBuildStateInfo{
			state: stateJobNotCreated,
		}
		return stateInfo, nil
	}

	stateInfo := &cBuildStateInfo{
		job: j,
	}

	finished, succeeded := jobStatus(j)
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
