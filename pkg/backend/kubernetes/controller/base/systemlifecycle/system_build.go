package systemlifecycle

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func isSystemBuildStatusCurrent(build *crv1.SystemBuild) bool {
	return build.Status.ObservedGeneration == build.Generation
}
