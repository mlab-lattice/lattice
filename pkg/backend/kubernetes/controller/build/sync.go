package build

import (
	"reflect"
	"sort"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncFailedBuild(build *latticev1.Build, stateInfo stateInfo) error {
	// Sort the ServiceBuild paths so the Status.Message is the same for the same failed ServiceBuilds
	var failedServices []tree.NodePath
	for service := range stateInfo.failedServiceBuilds {
		failedServices = append(failedServices, service)
	}

	sort.Slice(failedServices, func(i, j int) bool {
		return string(failedServices[i]) < string(failedServices[j])
	})

	message := "The following services failed to build:"
	for i, service := range failedServices {
		if i != 0 {
			message = message + ","
		}
		message = message + " " + string(service)
	}

	completionTimestamp := build.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateFailed,
		message,
		build.Status.StartTimestamp,
		completionTimestamp,
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncRunningBuild(build *latticev1.Build, stateInfo stateInfo) error {
	// Sort the ServiceBuild paths so the Status.Message is the same for the same failed ServiceBuilds
	var activeServices []tree.NodePath
	for service := range stateInfo.activeServiceBuilds {
		activeServices = append(activeServices, service)
	}

	sort.Slice(activeServices, func(i, j int) bool {
		return string(activeServices[i]) < string(activeServices[j])
	})

	message := "The following services are still building:"
	for i, service := range activeServices {
		if i != 0 {
			message = message + ","
		}
		message = message + " " + string(service)
	}

	// If we haven't logged a start timestamp yet, use now.
	// This should only happen if we created all of the service builds
	// but then failed to update the status.
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateRunning,
		message,
		build.Status.StartTimestamp,
		nil,
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncMissingServiceBuildsBuild(build *latticev1.Build, stateInfo stateInfo) error {
	// Copy so the shared cache isn't mutated
	status := build.Status.DeepCopy()
	serviceBuilds := status.ServiceBuilds
	if serviceBuilds == nil {
		serviceBuilds = map[tree.NodePath]string{}
	}

	serviceBuildStatuses := status.ServiceBuildStatuses
	if serviceBuildStatuses == nil {
		serviceBuildStatuses = map[string]latticev1.ServiceBuildStatus{}
	}

	for _, service := range stateInfo.needsNewServiceBuilds {
		serviceInfo := build.Spec.Services[service]

		// Otherwise we'll have to create a new Service.
		serviceBuild, err := c.createNewServiceBuild(build, service, serviceInfo.Definition)
		if err != nil {
			return err
		}

		serviceBuilds[service] = serviceBuild.Name
		serviceBuildStatuses[serviceBuild.Name] = serviceBuild.Status
	}

	// If we haven't logged a start timestamp yet, use now.
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateRunning,
		"",
		startTimestamp,
		nil,
		serviceBuilds,
		serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncSucceededBuild(build *latticev1.Build, stateInfo stateInfo) error {
	completionTimestamp := build.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateSucceeded,
		"",
		build.Status.StartTimestamp,
		build.Status.CompletionTimestamp,
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) updateSystemBuildStatus(
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

	return c.latticeClient.LatticeV1().Builds(build.Namespace).UpdateStatus(build)
}
