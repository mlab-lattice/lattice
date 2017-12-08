package componentbuild

import (
	"reflect"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"

	batchv1 "k8s.io/api/batch/v1"

	"github.com/golang/glog"
)

// Warning: syncJoblessComponentBuild mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *Controller) syncJoblessComponentBuild(cb *crv1.ComponentBuild) error {
	j, err := cbc.getBuildJob(cb)
	if err != nil {
		return err
	}

	jResp, err := cbc.kubeClient.BatchV1().Jobs(cb.Namespace).Create(j)
	if err != nil {
		// FIXME: send warn event
		return err
	}

	glog.V(4).Infof("Created Job %s", jResp.Name)
	// FIXME: send normal event
	return cbc.syncUnfinishedComponentBuild(cb, jResp)
}

// Warning: syncSuccessfulComponentBuild mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *Controller) syncSuccessfulComponentBuild(cb *crv1.ComponentBuild, j *batchv1.Job) error {
	newArtifacts := &crv1.ComponentBuildArtifacts{
		DockerImageFqn: j.Annotations[jobDockerFqnAnnotationKey],
	}

	if reflect.DeepEqual(cb.Status.State, crv1.ComponentBuildStateSucceeded) && reflect.DeepEqual(cb.Spec.Artifacts, newArtifacts) {
		return nil
	}

	cb.Status.State = crv1.ComponentBuildStateSucceeded
	cb.Spec.Artifacts = newArtifacts

	return cbc.putComponentBuildUpdate(cb)
}

// Warning: updateStatusToSucceeded mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *Controller) syncFailedComponentBuild(cb *crv1.ComponentBuild) error {
	return cbc.updateComponentBuildState(cb, crv1.ComponentBuildStateFailed)
}

// Warning: syncUnfinishedComponentBuild mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *Controller) syncUnfinishedComponentBuild(cb *crv1.ComponentBuild, j *batchv1.Job) error {
	// The Job Pods have been able to be scheduled, so the ComponentBuild is "running" even though
	// a Job Pod may not currently be active.
	if j.Status.Active > 0 || j.Status.Failed > 0 {
		return cbc.updateComponentBuildState(cb, crv1.ComponentBuildStateRunning)
	}

	return cbc.updateComponentBuildState(cb, crv1.ComponentBuildStateQueued)
}

// Warning: putComponentBuildStatusUpdate mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *Controller) updateComponentBuildState(cb *crv1.ComponentBuild, newState crv1.ComponentBuildState) error {
	if reflect.DeepEqual(cb.Status.State, newState) {
		return nil
	}

	cb.Status.State = newState
	return cbc.putComponentBuildUpdate(cb)
}

func (cbc *Controller) putComponentBuildUpdate(cb *crv1.ComponentBuild) error {
	_, err := cbc.latticeClient.LatticeV1().ComponentBuilds(cb.Namespace).Update(cb)
	return err
}
