package componentbuild

import (
	"reflect"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	batchv1 "k8s.io/api/batch/v1"

	"github.com/golang/glog"
)

// Warning: syncJoblessComponentBuild mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *ComponentBuildController) syncJoblessComponentBuild(cb *crv1.ComponentBuild) error {
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
func (cbc *ComponentBuildController) syncSuccessfulComponentBuild(cb *crv1.ComponentBuild, j *batchv1.Job) error {
	newStatus := crv1.ComponentBuildStatus{
		State: crv1.ComponentBuildStateSucceeded,
	}

	newArtifacts := &crv1.ComponentBuildArtifacts{
		DockerImageFqn: j.Annotations[jobDockerFqnAnnotationKey],
	}

	if reflect.DeepEqual(cb.Status, newStatus) && reflect.DeepEqual(cb.Spec.Artifacts, newArtifacts) {
		return nil
	}

	cb.Status = newStatus
	cb.Spec.Artifacts = newArtifacts

	return cbc.putComponentBuildUpdate(cb)
}

// Warning: updateStatusToSucceeded mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *ComponentBuildController) syncFailedComponentBuild(cb *crv1.ComponentBuild) error {
	newStatus := crv1.ComponentBuildStatus{
		State: crv1.ComponentBuildStateFailed,
		// TODO: add message explaining failure (from job logs or something?)
	}

	return cbc.putComponentBuildStatusUpdate(cb, newStatus)
}

// Warning: syncUnfinishedComponentBuild mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *ComponentBuildController) syncUnfinishedComponentBuild(cb *crv1.ComponentBuild, j *batchv1.Job) error {
	// The Job Pods have been able to be scheduled, so the ComponentBuild is "running" even though
	// a Job Pod may not currently be active.
	if j.Status.Active > 0 || j.Status.Failed > 0 {
		newStatus := crv1.ComponentBuildStatus{
			State: crv1.ComponentBuildStateRunning,
		}
		return cbc.putComponentBuildStatusUpdate(cb, newStatus)
	}

	// No Jobs have started executing, so we're still queued.
	newStatus := crv1.ComponentBuildStatus{
		State: crv1.ComponentBuildStateQueued,
	}
	return cbc.putComponentBuildStatusUpdate(cb, newStatus)
}

// Warning: putComponentBuildStatusUpdate mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *ComponentBuildController) putComponentBuildStatusUpdate(cb *crv1.ComponentBuild, newStatus crv1.ComponentBuildStatus) error {
	if reflect.DeepEqual(cb.Status, newStatus) {
		return nil
	}

	cb.Status = newStatus
	return cbc.putComponentBuildUpdate(cb)
}

func (cbc *ComponentBuildController) putComponentBuildUpdate(cb *crv1.ComponentBuild) error {
	err := cbc.latticeResourceClient.Put().
		Namespace(cb.Namespace).
		Resource(crv1.ComponentBuildResourcePlural).
		Name(cb.Name).
		Body(cb).
		Do().
		Into(nil)

	return err
}
