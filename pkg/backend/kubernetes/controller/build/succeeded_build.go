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

	completionTimestamp := build.Status.CompletionTimestamp
	if completionTimestamp == nil {
		now := metav1.Now()
		completionTimestamp = &now
	}

	_, err := c.updateBuildStatus(
		build,
		latticev1.BuildStateSucceeded,
		"",
		startTimestamp,
		completionTimestamp,
		build.Status.Services,
		build.Status.Jobs,
		stateInfo.containerBuildStatuses,
	)
	return err
}
