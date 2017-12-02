package constants

import (
	"github.com/mlab-lattice/system/pkg/types"
)

const (
	SystemRolloutStatePending    types.SystemRolloutState = "Pending"
	SystemRolloutStateAccepted   types.SystemRolloutState = "Accepted"
	SystemRolloutStateInProgress types.SystemRolloutState = "InProgress"
	SystemRolloutStateSucceeded  types.SystemRolloutState = "Succeeded"
	SystemRolloutStateFailed     types.SystemRolloutState = "Failed"
)
