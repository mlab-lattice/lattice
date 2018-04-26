package servicebuild

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) updateServiceBuildStatus(
	build *latticev1.ServiceBuild,
	state latticev1.ServiceBuildState,
	message string,
	startTimestamp *metav1.Time,
	completionTimestamp *metav1.Time,
	componentBuilds map[string]string,
	componentBuildStatuses map[string]latticev1.ComponentBuildStatus,
) (*latticev1.ServiceBuild, error) {
	status := latticev1.ServiceBuildStatus{
		State:   state,
		Message: message,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,

		ComponentBuilds:        componentBuilds,
		ComponentBuildStatuses: componentBuildStatuses,
	}

	if reflect.DeepEqual(build.Status, status) {
		return build, nil
	}

	// Copy so the shared cache isn't mutated
	build = build.DeepCopy()
	build.Status = status

	result, err := c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).UpdateStatus(build)
	if err != nil {
		return nil, fmt.Errorf("error updating status for %v: %v", build.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) deleteServiceBuild(build *latticev1.ServiceBuild) error {
	// don't attempt to delete dependants of the service build
	// the service build has a finalizer that will remove its ownerReferences
	// and then the component builds will sort themselves out
	orphanDelete := metav1.DeletePropagationOrphan
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &orphanDelete,
	}

	err := c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).Delete(build.Name, deleteOptions)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("error deleting %v: %v", build.Description(c.namespacePrefix), err)
	}

	return nil
}

func isOrphaned(build *latticev1.ServiceBuild) bool {
	return len(build.OwnerReferences) == 0
}

func (c *Controller) addFinalizer(build *latticev1.ServiceBuild) (*latticev1.ServiceBuild, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range build.Finalizers {
		if finalizer == kubeutil.ServiceBuildControllerFinalizer {
			return build, nil
		}
	}

	// Copy so we don't mutate the shared cache
	build = build.DeepCopy()
	build.Finalizers = append(build.Finalizers, kubeutil.AddressControllerFinalizer)

	result, err := c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).Update(build)
	if err != nil {
		return nil, fmt.Errorf("error adding %v finalizer: %v", build.Description(c.namespacePrefix), err)
	}

	return result, nil
}

func (c *Controller) removeFinalizer(build *latticev1.ServiceBuild) (*latticev1.ServiceBuild, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range build.Finalizers {
		if finalizer == kubeutil.ServiceBuildControllerFinalizer {
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

	result, err := c.latticeClient.LatticeV1().ServiceBuilds(build.Namespace).Update(build)
	if err != nil {
		return nil, fmt.Errorf("error removing %v finalizer: %v", build.Description(c.namespacePrefix), err)
	}

	return result, nil
}
