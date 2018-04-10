package systemlifecycle

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func isSystemBuildStatusCurrent(build *latticev1.Build) bool {
	return build.Status.ObservedGeneration == build.Generation
}
