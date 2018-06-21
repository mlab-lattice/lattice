package job

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"k8s.io/apimachinery/pkg/labels"
)

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
	jobRun.Finalizers = append(jobRun.Finalizers, kubeutil.ServiceControllerFinalizer)

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
