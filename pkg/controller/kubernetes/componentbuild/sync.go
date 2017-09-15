package componentbuild

import (
	"reflect"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	batchv1 "k8s.io/api/batch/v1"

	"github.com/golang/glog"
)

// Warning: syncJoblessComponentBuild mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *ComponentBuildController) syncJoblessComponentBuild(cBuild *crv1.ComponentBuild) error {
	job := cbc.getBuildJob(cBuild)
	jobResp, err := cbc.kubeClient.BatchV1().Jobs(cBuild.Namespace).Create(job)
	if err != nil {
		// FIXME: send warn event
		return err
	}

	glog.V(4).Infof("Created Job %s", jobResp.Name)
	// FIXME: send normal event
	return cbc.syncUnfinishedComponentBuild(cBuild, jobResp)
}

// Warning: syncSuccessfulComponentBuild mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *ComponentBuildController) syncSuccessfulComponentBuild(cBuild *crv1.ComponentBuild, job *batchv1.Job) error {
	newStatus := crv1.ComponentBuildStatus{
		State: crv1.ComponentBuildStateSucceeded,
	}

	newArtifacts := &crv1.ComponentBuildArtifacts{
		DockerImageFqn: job.Annotations[jobDockerFqnAnnotationKey],
	}

	if reflect.DeepEqual(cBuild.Status, newStatus) && reflect.DeepEqual(cBuild.Spec.Artifacts, newArtifacts) {
		return nil
	}

	return cbc.putComponentBuildUpdate(cBuild)
}

// Warning: updateStatusToSucceeded mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *ComponentBuildController) syncFailedComponentBuild(cBuild *crv1.ComponentBuild) error {
	newStatus := crv1.ComponentBuildStatus{
		State: crv1.ComponentBuildStateFailed,
		// TODO: add message explaining failure (from job logs or something?)
	}

	return cbc.putComponentBuildStatusUpdate(cBuild, newStatus)
}

// Warning: syncUnfinishedComponentBuild mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *ComponentBuildController) syncUnfinishedComponentBuild(cBuild *crv1.ComponentBuild, job *batchv1.Job) error {
	// The Job Pods have been able to be scheduled, so the ComponentBuild is "running" even though
	// a Job Pod may not currently be active.
	if job.Status.Active > 0 || job.Status.Failed > 0 {
		newStatus := crv1.ComponentBuildStatus{
			State: crv1.ComponentBuildStateRunning,
		}
		return cbc.putComponentBuildStatusUpdate(cBuild, newStatus)
	}

	// No Jobs have started executing, so we're still queued.
	newStatus := crv1.ComponentBuildStatus{
		State: crv1.ComponentBuildStateQueued,
	}
	return cbc.putComponentBuildStatusUpdate(cBuild, newStatus)
}

// Warning: putComponentBuildStatusUpdate mutates cBuild. Please do not pass in a pointer to a ComponentBuild
// from the shared cache.
func (cbc *ComponentBuildController) putComponentBuildStatusUpdate(cBuild *crv1.ComponentBuild, newStatus crv1.ComponentBuildStatus) error {
	if reflect.DeepEqual(cBuild.Status, newStatus) {
		return nil
	}

	cBuild.Status = newStatus
	return cbc.putComponentBuildUpdate(cBuild)
}

func (cbc *ComponentBuildController) putComponentBuildUpdate(cBuild *crv1.ComponentBuild) error {
	err := cbc.latticeResourceRestClient.Put().
		Namespace(cBuild.Namespace).
		Resource(crv1.ComponentBuildResourcePlural).
		Name(cBuild.Name).
		Body(cBuild).
		Do().
		Into(nil)

	return err
}
