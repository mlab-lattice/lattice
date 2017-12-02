package constants

import (
	"github.com/mlab-lattice/system/pkg/types"
)

const (
	ServiceBuildStatePending   types.ServiceBuildState = "Pending"
	ServiceBuildStateRunning   types.ServiceBuildState = "Running"
	ServiceBuildStateSucceeded types.ServiceBuildState = "Succeeded"
	ServiceBuildStateFailed    types.ServiceBuildState = "Failed"
)
