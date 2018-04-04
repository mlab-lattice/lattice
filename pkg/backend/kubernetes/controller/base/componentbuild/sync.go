package componentbuild

import (
	"encoding/json"
	"reflect"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	batchv1 "k8s.io/api/batch/v1"

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

func (c *Controller) syncSuccessfulComponentBuild(build *latticev1.ComponentBuild, j *batchv1.Job) error {
	artifacts := &latticev1.ComponentBuildArtifacts{
		DockerImageFQN: j.Annotations[jobDockerFqnAnnotationKey],
	}

	if reflect.DeepEqual(build.Status.State, latticev1.ComponentBuildStateSucceeded) && reflect.DeepEqual(build.Status.Artifacts, artifacts) {
		return nil
	}

	_, err := c.updateComponentBuildStatus(build, latticev1.ComponentBuildStateSucceeded, artifacts)
	return err
}

func (c *Controller) syncFailedComponentBuild(cb *latticev1.ComponentBuild) error {
	_, err := c.updateComponentBuildState(cb, latticev1.ComponentBuildStateFailed)
	return err
}

func (c *Controller) syncUnfinishedComponentBuild(cb *latticev1.ComponentBuild, j *batchv1.Job) error {
	// The Job Pods have been able to be scheduled, so the ComponentBuild is "running" even though
	// a Job Pod may not currently be active.
	if j.Status.Active > 0 || j.Status.Failed > 0 {
		_, err := c.updateComponentBuildState(cb, latticev1.ComponentBuildStateRunning)
		return err
	}

	_, err := c.updateComponentBuildState(cb, latticev1.ComponentBuildStateQueued)
	return err
}

func (c *Controller) updateComponentBuildState(
	build *latticev1.ComponentBuild,
	state latticev1.ComponentBuildState,
) (*latticev1.ComponentBuild, error) {
	return c.updateComponentBuildStatus(build, state, build.Status.Artifacts)
}

func (c *Controller) updateComponentBuildStatus(
	build *latticev1.ComponentBuild,
	state latticev1.ComponentBuildState,
	artifacts *latticev1.ComponentBuildArtifacts,
) (*latticev1.ComponentBuild, error) {
	var phasePtr *v1.ComponentBuildPhase
	if phase, ok := build.Annotations[kubeconstants.AnnotationKeyComponentBuildLastObservedPhase]; ok {
		phase := v1.ComponentBuildPhase(phase)
		phasePtr = &phase
	}

	var failureInfoPtr *v1.ComponentBuildFailureInfo
	if failureInfoData, ok := build.Annotations[kubeconstants.AnnotationKeyComponentBuildFailureInfo]; ok {
		failureInfo := v1.ComponentBuildFailureInfo{}
		err := json.Unmarshal([]byte(failureInfoData), &failureInfo)
		if err != nil {
			return nil, err
		}

		failureInfoPtr = &failureInfo
	}

	status := latticev1.ComponentBuildStatus{
		State:              state,
		ObservedGeneration: build.Generation,
		Artifacts:          artifacts,
		LastObservedPhase:  phasePtr,
		FailureInfo:        failureInfoPtr,
	}

	if reflect.DeepEqual(build.Status, status) {
		return build, nil
	}

	// Copy so the shared cache isn't mutated
	build = build.DeepCopy()
	build.Status = status
	return c.latticeClient.LatticeV1().ComponentBuilds(build.Namespace).Update(build)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().ComponentBuilds(build.Namespace).UpdateStatus(build)
}
