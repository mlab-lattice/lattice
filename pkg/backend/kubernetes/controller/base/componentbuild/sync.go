package componentbuild

import (
	"encoding/json"
	"reflect"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/types"

	batchv1 "k8s.io/api/batch/v1"

	"github.com/golang/glog"
)

func (c *Controller) syncJoblessComponentBuild(build *crv1.ComponentBuild) error {
	job, err := c.createNewJob(build)
	if err != nil {
		return err
	}

	glog.V(4).Infof("Created Job %s", job.Name)
	// FIXME: send normal event
	return c.syncUnfinishedComponentBuild(build, job)
}

func (c *Controller) syncSuccessfulComponentBuild(build *crv1.ComponentBuild, j *batchv1.Job) error {
	artifacts := &crv1.ComponentBuildArtifacts{
		DockerImageFQN: j.Annotations[jobDockerFqnAnnotationKey],
	}

	if reflect.DeepEqual(build.Status.State, crv1.ComponentBuildStateSucceeded) && reflect.DeepEqual(build.Status.Artifacts, artifacts) {
		return nil
	}

	_, err := c.updateComponentBuildStatus(build, crv1.ComponentBuildStateSucceeded, artifacts)
	return err
}

func (c *Controller) syncFailedComponentBuild(cb *crv1.ComponentBuild) error {
	_, err := c.updateComponentBuildState(cb, crv1.ComponentBuildStateFailed)
	return err
}

func (c *Controller) syncUnfinishedComponentBuild(cb *crv1.ComponentBuild, j *batchv1.Job) error {
	// The Job Pods have been able to be scheduled, so the ComponentBuild is "running" even though
	// a Job Pod may not currently be active.
	if j.Status.Active > 0 || j.Status.Failed > 0 {
		_, err := c.updateComponentBuildState(cb, crv1.ComponentBuildStateRunning)
		return err
	}

	_, err := c.updateComponentBuildState(cb, crv1.ComponentBuildStateQueued)
	return err
}

func (c *Controller) updateComponentBuildState(build *crv1.ComponentBuild, state crv1.ComponentBuildState) (*crv1.ComponentBuild, error) {
	return c.updateComponentBuildStatus(build, state, build.Status.Artifacts)
}

func (c *Controller) updateComponentBuildStatus(
	build *crv1.ComponentBuild,
	state crv1.ComponentBuildState,
	artifacts *crv1.ComponentBuildArtifacts,
) (*crv1.ComponentBuild, error) {
	var phasePtr *types.ComponentBuildPhase
	if phase, ok := build.Annotations[kubeconstants.AnnotationKeyComponentBuildLastObservedPhase]; ok {
		phase := types.ComponentBuildPhase(phase)
		phasePtr = &phase
	}

	var failureInfoPtr *types.ComponentBuildFailureInfo
	if failureInfoData, ok := build.Annotations[kubeconstants.AnnotationKeyComponentBuildFailureInfo]; ok {
		failureInfo := types.ComponentBuildFailureInfo{}
		err := json.Unmarshal([]byte(failureInfoData), &failureInfo)
		if err != nil {
			return nil, err
		}

		failureInfoPtr = &failureInfo
	}

	status := crv1.ComponentBuildStatus{
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
