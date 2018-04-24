package componentbuild

import (
	"fmt"
	"reflect"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
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

	glog.V(4).Infof("Created Job %s", job.Name)
	// FIXME: send normal event
	return c.syncUnfinishedComponentBuild(build, job)
}

func (c *Controller) syncSuccessfulComponentBuild(build *latticev1.ComponentBuild, job *batchv1.Job) error {
	dockerImageFQN, ok := build.DockerImageFQNAnnotation()
	if !ok {
		return fmt.Errorf(
			"%v claims to be in state %v but does not have %v annotation",
			build.Description(c.namespacePrefix),
			latticev1.ComponentBuildStateSucceeded,
			latticev1.ComponentBuildDockerImageFQNAnnotationKey,
		)
	}

	artifacts := &latticev1.ComponentBuildArtifacts{
		DockerImageFQN: dockerImageFQN,
	}

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
	// The Job Pods have been able to be scheduled, so the ComponentBuild is "running" even though
	// a Job Pod may not currently be active.
	if job.Status.Active > 0 || job.Status.Failed > 0 {
		startTimestamp := build.Status.StartTimestamp
		if startTimestamp == nil {
			now := metav1.Now()
			startTimestamp = &now
		}
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
		nil,
		nil,
		build.Status.Artifacts,
	)
	return err
}

func (c *Controller) updateComponentBuildState(
	build *latticev1.ComponentBuild,
	state latticev1.ComponentBuildState,
) (*latticev1.ComponentBuild, error) {
	startTimestamp := build.Status.StartTimestamp
	completionTimestamp := build.Status.CompletionTimestamp
	switch state {
	case latticev1.ComponentBuildStateFailed, latticev1.ComponentBuildStateSucceeded:

	case latticev1.ComponentBuildStateRunning:
		if startTimestamp == nil {
			now := metav1.Now()
			startTimestamp = &now
		}
	}

	return c.updateComponentBuildStatus(build, state, startTimestamp, completionTimestamp, build.Status.Artifacts)
}

func (c *Controller) updateComponentBuildStatus(
	build *latticev1.ComponentBuild,
	state latticev1.ComponentBuildState,
	startTimestamp *metav1.Time,
	completionTimestamp *metav1.Time,
	artifacts *latticev1.ComponentBuildArtifacts,
) (*latticev1.ComponentBuild, error) {
	var phasePtr *v1.ComponentBuildPhase
	if phase, ok := build.Annotations[kubeconstants.AnnotationKeyComponentBuildLastObservedPhase]; ok {
		phase := v1.ComponentBuildPhase(phase)
		phasePtr = &phase
	}

	failureInfo, err := build.FailureInfoAnnotation()
	if err != nil {
		return nil, err
	}

	status := latticev1.ComponentBuildStatus{
		State:       state,
		FailureInfo: failureInfo,

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
