package build

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) updateBuildStatus(
	build *latticev1.Build,
	state latticev1.BuildState,
	message string,
	startTimestamp *metav1.Time,
	completionTimestamp *metav1.Time,
	serviceBuilds map[tree.NodePath]string,
	serviceBuildStatuses map[string]latticev1.ServiceBuildStatus,
) (*latticev1.Build, error) {
	status := latticev1.BuildStatus{
		State:   state,
		Message: message,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,

		ServiceBuilds:        serviceBuilds,
		ServiceBuildStatuses: serviceBuildStatuses,
	}

	if reflect.DeepEqual(build.Status, status) {
		return build, nil
	}

	// Copy so the shared cache isn't mutated
	build = build.DeepCopy()
	build.Status = status

	result, err := c.latticeClient.LatticeV1().Builds(build.Namespace).UpdateStatus(build)
	if err != nil {
		return nil, fmt.Errorf("error updating status for %v: %v", build.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) addFinalizer(build *latticev1.Build) (*latticev1.Build, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range build.Finalizers {
		if finalizer == kubeutil.BuildControllerFinalizer {
			return build, nil
		}
	}

	// Copy so we don't mutate the shared cache
	build = build.DeepCopy()
	build.Finalizers = append(build.Finalizers, kubeutil.BuildControllerFinalizer)

	result, err := c.latticeClient.LatticeV1().Builds(build.Namespace).Update(build)
	if err != nil {
		return nil, fmt.Errorf("error adding %v finalizer: %v", build.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) removeFinalizer(build *latticev1.Build) (*latticev1.Build, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range build.Finalizers {
		if finalizer == kubeutil.BuildControllerFinalizer {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return build, nil
	}

	// Copy so we don't mutate the shared cache
	build = build.DeepCopy()
	build.Finalizers = finalizers

	result, err := c.latticeClient.LatticeV1().Builds(build.Namespace).Update(build)
	if err != nil {
		return nil, fmt.Errorf("error removing %v finalizer: %v", build.Description(c.namespacePrefix), err)
	}

	return result, nil
}
