package build

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncSucceededBuild(build *latticev1.Build, stateInfo stateInfo) error {
	startTimestamp := build.Status.StartTimestamp
	if startTimestamp == nil {
		now := metav1.Now()
		startTimestamp = &now
	}

	now := metav1.Now()
	_, err := c.updateBuildStatus(
		build,
		latticev1.BuildStateSucceeded,
		"",
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
