package build

import (
	"fmt"
	"sort"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type runningWorkloadInfo struct {
	MainContainerRunning bool
	RunningSidecars      []string
}

func (c *Controller) syncRunningBuild(build *latticev1.Build, stateInfo stateInfo) error {
	runningWorkloadsInfo := make(map[tree.Path]runningWorkloadInfo)
	var runningWorkloads []tree.Path

	for containerBuild := range stateInfo.activeContainerBuilds {
		workloads, ok := stateInfo.containerBuildWorkloads[containerBuild]
		if !ok {
			continue
		}

		runningWorkloads = append(runningWorkloads, workloads...)

		for _, path := range workloads {
			workloadInfo, ok := build.Status.Workloads[path]
			if !ok {
				// this really really shouldn't happen...
				runningWorkloadsInfo[path] = runningWorkloadInfo{}
				continue
			}

			info := runningWorkloadInfo{}
			if workloadInfo.MainContainer == containerBuild {
				info.MainContainerRunning = true
			}

			for sidecar, sidecarBuild := range workloadInfo.Sidecars {
				if sidecarBuild == containerBuild {
					info.RunningSidecars = append(info.RunningSidecars, sidecar)
				}
			}

			runningWorkloadsInfo[path] = info
		}
	}

	// Sort the workloads paths so the message is the same for the same failed container builds
	sort.Slice(runningWorkloads, func(i, j int) bool {
		return string(runningWorkloads[i]) < string(runningWorkloads[j])
	})

	message := "the following workloads are still building: "
	for i, path := range runningWorkloads {
		if i != 0 {
			message = message + ","
		}

		info := runningWorkloadsInfo[path]
		serviceMessage := fmt.Sprintf("%v (", path.String())
		previousFailure := false

		if info.MainContainerRunning {
			serviceMessage += "main container"
			previousFailure = true
		}

		for _, sidecar := range info.RunningSidecars {
			if previousFailure {
				serviceMessage += ", "
			}

			serviceMessage += fmt.Sprintf("%v sidecar", sidecar)
		}

		message = message + " " + serviceMessage
	}

	// If we haven't logged a start timestamp yet, use now.
	// This should only happen if we created all of the path builds
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
		nil,
		build.Status.Definition,
		startTimestamp,
		nil,
		build.Status.Workloads,
		stateInfo.containerBuildStatuses,
	)
	return err
}
