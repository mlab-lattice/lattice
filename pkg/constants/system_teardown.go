package constants

import (
	"github.com/mlab-lattice/system/pkg/types"
)

const (
	SystemTeardownStatePending    types.SystemTeardownState = "Pending"
	SystemTeardownStateInProgress types.SystemTeardownState = "InProgress"
	SystemTeardownStateSucceeded  types.SystemTeardownState = "Succeeded"
	SystemTeardownStateFailed     types.SystemTeardownState = "Failed"
)
