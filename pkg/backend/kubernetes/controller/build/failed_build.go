package build

import (
	"fmt"
	"sort"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type failedServiceInfo struct {
	MainContainerFailed bool
	FailedSidecars      []string
}

func (c *Controller) syncFailedBuild(build *latticev1.Build, stateInfo stateInfo) error {
	failedServicesInfo := make(map[tree.NodePath]failedServiceInfo)
	failedJobsInfo := make(map[tree.NodePath]failedServiceInfo)
	var failedServices []tree.NodePath
	var failedJobs []tree.NodePath

	for buildName := range stateInfo.failedContainerBuilds {
		buildFailedServices, ok := stateInfo.containerBuildServices[buildName]
		if !ok {
			continue
		}

		failedServices = append(failedServices, buildFailedServices...)

		for _, servicePath := range buildFailedServices {
			serviceInfo, ok := build.Status.Services[servicePath]
			if !ok {
				// this really really shouldn't happen...
				failedServicesInfo[servicePath] = failedServiceInfo{}
				continue
			}

			info := failedServiceInfo{}
			if serviceInfo.MainContainer == buildName {
				info.MainContainerFailed = true
			}

			for sidecar, sidecarBuild := range serviceInfo.Sidecars {
				if sidecarBuild == buildName {
					info.FailedSidecars = append(info.FailedSidecars, sidecar)
				}
			}

			failedServicesInfo[servicePath] = info
		}

		buildFailedJobs, ok := stateInfo.containerBuildJobs[buildName]
		if !ok {
			continue
		}

		failedJobs = append(failedJobs, buildFailedJobs...)

		for _, jobPath := range buildFailedJobs {
			jobInfo, ok := build.Status.Jobs[jobPath]
			if !ok {
				// this really really shouldn't happen...
				failedJobsInfo[jobPath] = failedServiceInfo{}
				continue
			}

			info := failedServiceInfo{}
			if jobInfo.MainContainer == buildName {
				info.MainContainerFailed = true
			}

			for sidecar, sidecarBuild := range jobInfo.Sidecars {
				if sidecarBuild == buildName {
					info.FailedSidecars = append(info.FailedSidecars, sidecar)
				}
			}

			failedJobsInfo[jobPath] = info
		}
	}

	// Sort the service paths so the message is the same for the same failed container builds
	sort.Slice(failedServices, func(i, j int) bool {
		return string(failedServices[i]) < string(failedServices[j])
	})

	message := "the following services failed to build:"
	for i, service := range failedServices {
		if i != 0 {
			message = message + ","
		}

		info := failedServicesInfo[service]
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

	// Sort the job paths so the message is the same for the same failed container builds
	sort.Slice(failedJobs, func(i, j int) bool {
		return string(failedJobs[i]) < string(failedJobs[j])
	})

	message += "\nthe following jobs failed to build:"
	for i, job := range failedJobs {
		if i != 0 {
			message = message + ","
		}

		info := failedJobsInfo[job]
		serviceMessage := fmt.Sprintf("%v (", job.String())
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

	// if we haven't logged a start timestamp yet, use now
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	// if we haven't logged a completion timestamp yet, use now
	completionTimestamp := build.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	_, err := c.updateBuildStatus(
		build,
		latticev1.BuildStateFailed,
		message,
		startTimestamp,
		completionTimestamp,
		build.Status.Services,
		build.Status.Jobs,
		stateInfo.containerBuildStatuses,
	)
	return err
}
