package build

import (
	"fmt"
	"sort"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type failedWorkloadInfo struct {
	MainContainerFailed bool
	FailedSidecars      []string
}

func (c *Controller) syncFailedBuild(build *latticev1.Build, stateInfo stateInfo) error {
	failedWorkloadsInfo := make(map[tree.Path]failedWorkloadInfo)
	var failedWorkloads []tree.Path

	// collect information about which workloads have failed to build given the failed
	// container builds
	for containerBuild := range stateInfo.failedContainerBuilds {
		containerBuildWorkloads, ok := stateInfo.containerBuildWorkloads[containerBuild]
		if !ok {
			continue
		}

		failedWorkloads = append(failedWorkloads, containerBuildWorkloads...)

		for _, path := range containerBuildWorkloads {
			workloadInfo, ok := build.Status.Workloads[path]
			if !ok {
				// this really shouldn't happen...
				failedWorkloadsInfo[path] = failedWorkloadInfo{}
				continue
			}

			info := failedWorkloadInfo{}
			if workloadInfo.MainContainer == containerBuild {
				info.MainContainerFailed = true
			}

			for sidecar, sidecarBuild := range workloadInfo.Sidecars {
				if sidecarBuild == containerBuild {
					info.FailedSidecars = append(info.FailedSidecars, sidecar)
				}
			}

			failedWorkloadsInfo[path] = info
		}
	}

	// sort the workload paths so the message is the same for the same failed container builds
	sort.Slice(failedWorkloads, func(i, j int) bool {
		return string(failedWorkloads[i]) < string(failedWorkloads[j])
	})

	message := "the following workloads failed to build:"
	for i, workload := range failedWorkloads {
		if i != 0 {
			message = message + ","
		}

		info := failedWorkloadsInfo[workload]
		serviceMessage := fmt.Sprintf("%v (", workload.String())
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

	now := metav1.Now()
	_, err := c.updateBuildStatus(
		build,
		latticev1.BuildStateFailed,
		message,
		nil,
		build.Status.Definition,
		build.Status.Path,
		build.Status.Version,
		build.Status.StartTimestamp,
		&now,
		build.Status.Workloads,
		stateInfo.containerBuildStatuses,
	)
	return err
}
