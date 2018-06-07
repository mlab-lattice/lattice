package build

import (
	"fmt"
	"sort"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type runningServiceInfo struct {
	MainContainerFailed bool
	FailedSidecars      []string
}

func (c *Controller) syncRunningBuild(build *latticev1.Build, stateInfo stateInfo) error {
	runningServicesInfo := make(map[tree.NodePath]runningServiceInfo)
	var activeServices []tree.NodePath

	for buildName := range stateInfo.activeContainerBuilds {
		buildActiveServices, ok := stateInfo.containerBuildServices[buildName]
		if !ok {
			continue
		}

		activeServices = append(activeServices, buildActiveServices...)

		for _, servicePath := range buildActiveServices {
			serviceInfo, ok := build.Status.Services[servicePath]
			if !ok {
				// this really really shouldn't happen...
				runningServicesInfo[servicePath] = runningServiceInfo{}
				continue
			}

			info := runningServiceInfo{}
			if serviceInfo.MainContainer == buildName {
				info.MainContainerFailed = true
			}

			for sidecar, sidecarBuild := range serviceInfo.Sidecars {
				if sidecarBuild == buildName {
					info.FailedSidecars = append(info.FailedSidecars, sidecar)
				}
			}
		}
	}

	// Sort the service paths so the message is the same for the same failed container builds
	sort.Slice(activeServices, func(i, j int) bool {
		return string(activeServices[i]) < string(activeServices[j])
	})

	message := "the following services are still building:"
	for i, service := range activeServices {
		if i != 0 {
			message = message + ","
		}

		info := runningServicesInfo[service]
		serviceMessage := fmt.Sprintf("%v (", service.String())
		previousFailure := false

		if info.MainContainerFailed {
			serviceMessage += "main container"
			previousFailure = true
		}

		for _, sidecar := range info.FailedSidecars {
			if previousFailure {
				serviceMessage += ", "
			}

			serviceMessage += fmt.Sprintf("%v sidecar", sidecar)
		}

		message = message + " " + serviceMessage
	}

	// If we haven't logged a start timestamp yet, use now.
	// This should only happen if we created all of the service builds
	// but then failed to update the status.
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	_, err := c.updateBuildStatus(
		build,
		latticev1.BuildStateRunning,
		message,
		startTimestamp,
		nil,
		build.Status.Services,
		stateInfo.containerBuildStatuses,
	)
	return err
}
