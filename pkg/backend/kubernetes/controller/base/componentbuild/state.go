package componentbuild

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	batchv1 "k8s.io/api/batch/v1"
)

type cBuildState string

const (
	cBuildStateJobNotCreated cBuildState = "job-not-created"
	cBuildStateJobRunning    cBuildState = "job-running"
	cBuildStateJobSucceeded  cBuildState = "job-succeeded"
	cBuildStateJobFailed     cBuildState = "job-failed"
)

type cBuildStateInfo struct {
	state cBuildState
	job   *batchv1.Job
}

func (cbc *Controller) calculateState(cb *crv1.ComponentBuild) (*cBuildStateInfo, error) {
	j, err := cbc.getJobForBuild(cb)
	if err != nil {
		return nil, err
	}

	// FIXME: if a ComponentBuild was successful, but then for some reason the Job is deleted, should it still be
	// considered successful or should a new Job be spun up? Right now a new Job will be spun up.
	if j == nil {
		stateInfo := &cBuildStateInfo{
			state: cBuildStateJobNotCreated,
		}
		return stateInfo, nil
	}

	stateInfo := &cBuildStateInfo{
		job: j,
	}

	finished, succeeded := jobStatus(j)
	if !finished {
		stateInfo.state = cBuildStateJobRunning
		return stateInfo, nil
	}

	stateInfo.state = cBuildStateJobSucceeded
	if !succeeded {
		stateInfo.state = cBuildStateJobFailed
	}

	return stateInfo, nil
}
