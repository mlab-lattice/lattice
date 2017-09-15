package componentbuild

import (
	"fmt"
	"reflect"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	batchv1 "k8s.io/api/batch/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type cBuildState string

const (
	cBuildStateJobNotCreated  cBuildState = "job-not-created"
	cBuildStateJobNotFinished cBuildState = "job-not-finished"
	cBuildStateJobSucceeded   cBuildState = "job-succeeded"
	cBuildStateJobFailed      cBuildState = "job-failed"
)

type cBuildStateInfo struct {
	state cBuildState
	job   *batchv1.Job
}

func (cbc *ComponentBuildController) calculateState(cBuild *crv1.ComponentBuild) (*cBuildStateInfo, error) {
	job, err := cbc.getJobForBuild(cBuild)
	if err != nil {
		return nil, err
	}

	if job == nil {
		stateInfo := &cBuildStateInfo{
			state: cBuildStateJobNotCreated,
		}
		return stateInfo, nil
	}

	stateInfo := &cBuildStateInfo{
		job: job,
	}

	finished, succeeded := jobStatus(job)
	if !finished {
		stateInfo.state = cBuildStateJobNotFinished
		return stateInfo, nil
	}

	stateInfo.state = cBuildStateJobSucceeded
	if !succeeded {
		stateInfo.state = cBuildStateJobFailed
	}

	return stateInfo, nil
}

// getJobForBuild uses ControllerRefManager to retrieve the Job for a ComponentBuild
func (cbc *ComponentBuildController) getJobForBuild(cBuild *crv1.ComponentBuild) (*batchv1.Job, error) {
	// List all Jobs to find in the ComponentBuild's namespace to find the Job the ComponentBuild manages.
	jobList, err := cbc.jobLister.Jobs(cBuild.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	matchingJobs := []*batchv1.Job{}
	cBuildControllerRef := metav1.NewControllerRef(cBuild, controllerKind)

	for _, job := range jobList {
		jobControllerRef := metav1.GetControllerOf(job)

		if reflect.DeepEqual(cBuildControllerRef, jobControllerRef) {
			matchingJobs = append(matchingJobs, job)
		}
	}

	if len(matchingJobs) == 0 {
		return nil, nil
	}

	if len(matchingJobs) > 1 {
		return nil, fmt.Errorf("ComponentBuild %v has multiple Jobs", cBuild.Name)
	}

	return matchingJobs[0], nil
}
