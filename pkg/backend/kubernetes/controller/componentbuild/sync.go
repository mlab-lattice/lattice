package componentbuild

import (
	"fmt"
	"reflect"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	batchv1 "k8s.io/api/batch/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncJoblessComponentBuild(build *latticev1.ComponentBuild) error {
	job, err := c.createNewJob(build)
	if err != nil {
		return err
	}

	glog.V(4).Infof("created job %v for %v", job.Name, build.Description(c.namespacePrefix))
	return c.syncUnfinishedComponentBuild(build, job)
}

func (c *Controller) syncSuccessfulComponentBuild(build *latticev1.ComponentBuild, job *batchv1.Job) error {
	dockerImageFQN, ok := job.Annotations[latticev1.ComponentBuildJobDockerImageFQNAnnotationKey]
	if !ok {
		return fmt.Errorf(
			"job %v for %v claims to have succeeded but does not have %v annotation",
			job.Name,
			build.Description(c.namespacePrefix),
			latticev1.ComponentBuildJobDockerImageFQNAnnotationKey,
		)
	}

	artifacts := &latticev1.ComponentBuildArtifacts{
		DockerImageFQN: dockerImageFQN,
	}

	// if we haven't logged a start timestamp yet, use now
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	// if we haven't logged a completion timestamp yet, use now
	completionTimestamp := build.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	_, err := c.updateComponentBuildStatus(
		build,
		latticev1.ComponentBuildStateSucceeded,
		build.Status.StartTimestamp,
		completionTimestamp,
		artifacts,
	)
	return err
}

func (c *Controller) syncFailedComponentBuild(build *latticev1.ComponentBuild) error {
	// if we haven't logged a start timestamp yet, use now
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	// if we haven't logged a start timestamp yet, use now
	completionTimestamp := build.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	_, err := c.updateComponentBuildStatus(
		build,
		latticev1.ComponentBuildStateFailed,
		build.Status.StartTimestamp,
		completionTimestamp,
		build.Status.Artifacts,
	)
	return err
}

func (c *Controller) syncUnfinishedComponentBuild(build *latticev1.ComponentBuild, job *batchv1.Job) error {
	// if we haven't logged a start timestamp yet, use now
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	if job.Status.Active > 0 || job.Status.Failed > 0 {
		_, err := c.updateComponentBuildStatus(
			build,
			latticev1.ComponentBuildStateRunning,
			startTimestamp,
			nil,
			build.Status.Artifacts,
		)
		return err
	}

	_, err := c.updateComponentBuildStatus(
		build,
		latticev1.ComponentBuildStateQueued,
		startTimestamp,
		nil,
		build.Status.Artifacts,
	)
	return err
}

func (c *Controller) updateComponentBuildStatus(
	build *latticev1.ComponentBuild,
	state latticev1.ComponentBuildState,
	startTimestamp *metav1.Time,
	completionTimestamp *metav1.Time,
	artifacts *latticev1.ComponentBuildArtifacts,
) (*latticev1.ComponentBuild, error) {
	var phasePtr *v1.ComponentBuildPhase
	if phase, ok := build.LastObservedPhaseAnnotation(); ok {
		phasePtr = &phase
	}

	failureInfo, err := build.FailureInfoAnnotation()
	if err != nil {
		return nil, err
	}

	status := latticev1.ComponentBuildStatus{
		State:       state,
		FailureInfo: failureInfo,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,

		Artifacts:         artifacts,
		LastObservedPhase: phasePtr,
	}

	if reflect.DeepEqual(build.Status, status) {
		return build, nil
	}

	// Copy so the shared cache isn't mutated
	build = build.DeepCopy()
	build.Status = status

	return c.latticeClient.LatticeV1().ComponentBuilds(build.Namespace).UpdateStatus(build)
}
