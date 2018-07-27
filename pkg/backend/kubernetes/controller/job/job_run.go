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

func (c *Controller) syncJobRunStatus(jobRun *latticev1.JobRun, kubeJob *batchv1.Job) (*latticev1.JobRun, error) {
	startTimestamp := jobRun.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	completionTimestamp := jobRun.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	state := jobRunState(kubeJob.Status)
	return c.updateJobRunStatus(jobRun, state, startTimestamp, completionTimestamp)
}

func jobRunState(kubeJobStatus batchv1.JobStatus) latticev1.JobRunState {
	if kubeJobStatusContainsCondition(kubeJobStatus, batchv1.JobComplete) {
		return latticev1.JobRunStateSucceeded
	}

	if kubeJobStatusContainsCondition(kubeJobStatus, batchv1.JobFailed) {
		return latticev1.JobRunStateFailed
	}

	if kubeJobStatus.Active > 0 || kubeJobStatus.Failed > 0 || kubeJobStatus.Succeeded > 0 {
		return latticev1.JobRunStateRunning
	}

	return latticev1.JobRunStateQueued
}

func kubeJobStatusContainsCondition(kubeJobStatus batchv1.JobStatus, conditionType batchv1.JobConditionType) bool {
	for _, c := range kubeJobStatus.Conditions {
		if c.Type == conditionType {
			return c.Status == corev1.ConditionTrue
		}
	}

	return false
}

func (c *Controller) updateJobRunStatus(
	jobRun *latticev1.JobRun,
	state latticev1.JobRunState,
	startTimestamp, completionTimestamp *metav1.Time,
) (*latticev1.JobRun, error) {
	status := latticev1.JobRunStatus{
		State:               state,
		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,
	}

	if reflect.DeepEqual(jobRun.Status, status) {
		return jobRun, nil
	}

	// copy so we don't mutate the cache
	jobRun = jobRun.DeepCopy()
	jobRun.Status = status

	result, err := c.latticeClient.LatticeV1().JobRuns(jobRun.Namespace).UpdateStatus(jobRun)
	if err != nil {
		return nil, fmt.Errorf("error updating status for %v: %v", jobRun.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) nodePoolJobRuns(nodePool *latticev1.NodePool) ([]latticev1.JobRun, error) {
	_, ok, err := nodePool.SystemSharedPathLabel()
	if err != nil {
		return nil, err
	}
	if ok {
		jobRuns, err := c.jobRunLister.JobRuns(nodePool.Namespace).List(labels.Everything())
		if err != nil {
			return nil, err
		}

		var nodePoolJobRuns []latticev1.JobRun
		for _, jobRun := range jobRuns {
			// FIXME: this method was not working for services that had not yet annotated themselves
			//nodePools, err := jobRun.NodePoolAnnotation()
			//if err != nil {
			//	continue
			//}

			//if nodePools.ContainsNodePool(nodePool.Namespace, nodePool.Name) {
			nodePoolJobRuns = append(nodePoolJobRuns, *jobRun)
			//}
		}
		return nodePoolJobRuns, nil
	}

	err = fmt.Errorf(
		"%v did not have %v annotation",
		nodePool.Description(c.namespacePrefix),
		latticev1.NodePoolSystemSharedPathLabelKey,
	)
	return nil, err
}

func (c *Controller) addFinalizer(jobRun *latticev1.JobRun) (*latticev1.JobRun, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range jobRun.Finalizers {
		if finalizer == kubeutil.JobControllerFinalizer {
			return jobRun, nil
		}
	}

	// Copy so we don't mutate the shared cache
	jobRun = jobRun.DeepCopy()
	jobRun.Finalizers = append(jobRun.Finalizers, kubeutil.JobControllerFinalizer)

	result, err := c.latticeClient.LatticeV1().JobRuns(jobRun.Namespace).Update(jobRun)
	if err != nil {
		return nil, fmt.Errorf("error adding %v finalizer: %v", jobRun.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) removeFinalizer(jobRun *latticev1.JobRun) (*latticev1.JobRun, error) {
	var finalizers []string
	found := false
	for _, finalizer := range jobRun.Finalizers {
		if finalizer == kubeutil.JobControllerFinalizer {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return jobRun, nil
	}

	// Copy so we don't mutate the shared cache
	jobRun = jobRun.DeepCopy()
	jobRun.Finalizers = finalizers

	result, err := c.latticeClient.LatticeV1().JobRuns(jobRun.Namespace).Update(jobRun)
	if err != nil {
		return nil, fmt.Errorf("error removing %v finalizer: %v", jobRun.Description(c.namespacePrefix), err)
	}

	return result, nil
}
