package build

import (
	"fmt"
	"reflect"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) updateBuildStatus(
	build *latticev1.Build,
	state latticev1.BuildState,
	definition *definitionv1.SystemNode,
	resolutionInfo resolver.ResolutionInfo,
	message string,
	startTimestamp *metav1.Time,
	completionTimestamp *metav1.Time,
	services map[tree.Path]latticev1.BuildStatusService,
	jobs map[tree.Path]latticev1.BuildStatusJob,
	containerBuildStatuses map[string]latticev1.ContainerBuildStatus,
) (*latticev1.Build, error) {
	status := latticev1.BuildStatus{
		State:   state,
		Message: message,

		Definition:     definition,
		ResolutionInfo: resolutionInfo,

		StartTimestamp:      startTimestamp,
		CompletionTimestamp: completionTimestamp,

		Services: services,
		Jobs:     jobs,
		ContainerBuildStatuses: containerBuildStatuses,
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
