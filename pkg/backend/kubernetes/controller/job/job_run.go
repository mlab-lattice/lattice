package job

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
)

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
