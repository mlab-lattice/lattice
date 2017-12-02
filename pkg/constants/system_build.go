package constants

import (
	"github.com/mlab-lattice/system/pkg/types"
)

const (
	SystemBuildStatePending   types.SystemBuildState = "Pending"
	SystemBuildStateRunning   types.SystemBuildState = "Running"
	SystemBuildStateSucceeded types.SystemBuildState = "Succeeded"
	SystemBuildStateFailed    types.SystemBuildState = "Failed"
)
