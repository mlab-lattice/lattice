package build

import (
	"fmt"
	"sort"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
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
		serviceMessage := path.String()

		hasSidecars := len(info.RunningSidecars) != 0
		if hasSidecars {
			serviceMessage += "("
		}

		delim := ""
		if info.MainContainerRunning && hasSidecars {
			serviceMessage += "main container"
			delim = ", "
		}

		for _, sidecar := range info.RunningSidecars {
			serviceMessage += delim
			serviceMessage += fmt.Sprintf("%v sidecar", sidecar)
			delim = ", "
		}

		message += serviceMessage
		if hasSidecars {
			message += ")"
		}
	}

	_, err := c.updateBuildStatus(
		build,
		latticev1.BuildStateRunning,
		message,
		nil,
		build.Status.Definition,
		build.Status.Path,
		build.Status.Version,
		build.Status.StartTimestamp,
		nil,
		build.Status.Workloads,
		stateInfo.containerBuildStatuses,
	)
	return err
}
