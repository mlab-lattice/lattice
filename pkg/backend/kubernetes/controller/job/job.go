package job

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) syncJobStatus(job *latticev1.Job, kubeJob *batchv1.Job) (*latticev1.Job, error) {
	startTimestamp := job.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	completionTimestamp := job.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	state := jobState(kubeJob.Status)
	return c.updateJobStatus(job, state, startTimestamp, completionTimestamp)
}

func jobState(kubeJobStatus batchv1.JobStatus) latticev1.JobState {
	if kubeJobStatusContainsCondition(kubeJobStatus, batchv1.JobComplete) {
		return latticev1.JobStateSucceeded
	}

	if kubeJobStatusContainsCondition(kubeJobStatus, batchv1.JobFailed) {
		return latticev1.JobStateFailed
	}

	if kubeJobStatus.Active > 0 || kubeJobStatus.Failed > 0 || kubeJobStatus.Succeeded > 0 {
		return latticev1.JobStateRunning
	}

	return latticev1.JobStateQueued
}

func kubeJobStatusContainsCondition(kubeJobStatus batchv1.JobStatus, conditionType batchv1.JobConditionType) bool {
	for _, c := range kubeJobStatus.Conditions {
		if c.Type == conditionType {
			return c.Status == corev1.ConditionTrue
		}
	}

	return false
}

func (c *Controller) updateJobStatus(
	job *latticev1.Job,
	state latticev1.JobState,
	startTimestamp, completionTimestamp *metav1.Time,
) (*latticev1.Job, error) {
	status := latticev1.JobStatus{
		State:               state,
		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,
	}

	if reflect.DeepEqual(job.Status, status) {
		return job, nil
	}

	// copy so we don't mutate the cache
	job = job.DeepCopy()
	job.Status = status

	result, err := c.latticeClient.LatticeV1().Jobs(job.Namespace).UpdateStatus(job)
	if err != nil {
		return nil, fmt.Errorf("error updating status for %v: %v", job.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) nodePoolJobs(nodePool *latticev1.NodePool) ([]latticev1.Job, error) {
	_, ok, err := nodePool.SystemSharedPathLabel()
	if err != nil {
		return nil, err
	}
	if ok {
		jobs, err := c.jobLister.Jobs(nodePool.Namespace).List(labels.Everything())
		if err != nil {
			return nil, err
		}

		var nodePoolJobs []latticev1.Job
		for _, job := range jobs {
			// FIXME: this method was not working for services that had not yet annotated themselves
			//nodePools, err := job.NodePoolAnnotation()
			//if err != nil {
			//	continue
			//}

			//if nodePools.ContainsNodePool(nodePool.Namespace, nodePool.Name) {
			nodePoolJobs = append(nodePoolJobs, *job)
			//}
		}
		return nodePoolJobs, nil
	}

	err = fmt.Errorf(
		"%v did not have %v annotation",
		nodePool.Description(c.namespacePrefix),
		latticev1.NodePoolSystemSharedPathLabelKey,
	)
	return nil, err
}

func (c *Controller) addFinalizer(job *latticev1.Job) (*latticev1.Job, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range job.Finalizers {
		if finalizer == kubeutil.JobControllerFinalizer {
			return job, nil
		}
	}

	// Copy so we don't mutate the shared cache
	job = job.DeepCopy()
	job.Finalizers = append(job.Finalizers, kubeutil.JobControllerFinalizer)

	result, err := c.latticeClient.LatticeV1().Jobs(job.Namespace).Update(job)
	if err != nil {
		return nil, fmt.Errorf("error adding %v finalizer: %v", job.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) removeFinalizer(job *latticev1.Job) (*latticev1.Job, error) {
	var finalizers []string
	found := false
	for _, finalizer := range job.Finalizers {
		if finalizer == kubeutil.JobControllerFinalizer {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return job, nil
	}

	// Copy so we don't mutate the shared cache
	job = job.DeepCopy()
	job.Finalizers = finalizers

	result, err := c.latticeClient.LatticeV1().Jobs(job.Namespace).Update(job)
	if err != nil {
		return nil, fmt.Errorf("error removing %v finalizer: %v", job.Description(c.namespacePrefix), err)
	}

	return result, nil
}
