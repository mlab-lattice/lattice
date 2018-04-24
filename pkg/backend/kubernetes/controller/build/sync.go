package build

import (
	"reflect"
	"sort"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func (c *Controller) syncFailedSystemBuild(build *latticev1.Build, stateInfo stateInfo) error {
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

	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateFailed,
		message,
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncRunningSystemBuild(build *latticev1.Build, stateInfo stateInfo) error {
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

	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateRunning,
		message,
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncMissingServiceBuildsSystemBuild(build *latticev1.Build, stateInfo stateInfo) error {
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

	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateRunning,
		"",
		serviceBuilds,
		serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncSucceededSystemBuild(build *latticev1.Build, stateInfo stateInfo) error {
	_, err := c.updateSystemBuildStatus(
		build,
		latticev1.BuildStateSucceeded,
		"",
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) updateSystemBuildStatus(
	build *latticev1.Build,
	state latticev1.BuildState,
	message string,
	serviceBuilds map[tree.NodePath]string,
	serviceBuildStatuses map[string]latticev1.ServiceBuildStatus,
) (*latticev1.Build, error) {
	status := latticev1.BuildStatus{
		State:                state,
		Message:              message,
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
