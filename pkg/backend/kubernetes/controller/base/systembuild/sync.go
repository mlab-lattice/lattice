package systembuild

import (
	"reflect"
	"sort"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

func (c *Controller) syncFailedSystemBuild(build *crv1.SystemBuild, stateInfo stateInfo) error {
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
		crv1.SystemBuildStateFailed,
		message,
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncRunningSystemBuild(build *crv1.SystemBuild, stateInfo stateInfo) error {
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
		crv1.SystemBuildStateRunning,
		message,
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncMissingServiceBuildsSystemBuild(build *crv1.SystemBuild, stateInfo stateInfo) error {
	// Copy so the shared cache isn't mutated
	status := build.Status.DeepCopy()
	serviceBuilds := status.ServiceBuilds
	serviceBuildStatuses := status.ServiceBuildStatuses

	for _, service := range stateInfo.needsNewServiceBuilds {
		serviceInfo := build.Spec.Services[service]

		// Otherwise we'll have to create a new Service.
		serviceBuild, err := c.createNewServiceBuild(build, service, &serviceInfo.Definition)
		if err != nil {
			return err
		}

		serviceBuilds[service] = serviceBuild.Name
		serviceBuildStatuses[serviceBuild.Name] = serviceBuild.Status
	}

	_, err := c.updateSystemBuildStatus(
		build,
		crv1.SystemBuildStateRunning,
		"",
		serviceBuilds,
		serviceBuildStatuses,
	)
	return err
}

func (c *Controller) syncSucceededSystemBuild(build *crv1.SystemBuild, stateInfo stateInfo) error {
	_, err := c.updateSystemBuildStatus(
		build,
		crv1.SystemBuildStateSucceeded,
		"",
		stateInfo.serviceBuilds,
		stateInfo.serviceBuildStatuses,
	)
	return err
}

func (c *Controller) putSystemBuildUpdate(sysb *crv1.SystemBuild) (*crv1.SystemBuild, error) {
	return c.latticeClient.LatticeV1().SystemBuilds(sysb.Namespace).Update(sysb)
}

func (c *Controller) updateSystemBuildStatus(
	build *crv1.SystemBuild,
	state crv1.SystemBuildState,
	message string,
	serviceBuilds map[tree.NodePath]string,
	serviceBuildStatuses map[string]crv1.ServiceBuildStatus,
) (*crv1.SystemBuild, error) {
	status := crv1.SystemBuildStatus{
		State:                state,
		ObservedGeneration:   build.Generation,
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
	return c.latticeClient.LatticeV1().SystemBuilds(build.Namespace).Update(build)
}
