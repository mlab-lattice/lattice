package constants

import (
	"github.com/mlab-lattice/system/pkg/types"
)

const (
	ComponentBuildPhasePullingGitRepository types.ComponentBuildPhase = "pulling git repository"
	ComponentBuildPhaseBuildingDockerImage  types.ComponentBuildPhase = "building docker image"
	ComponentBuildPhasePushingDockerImage   types.ComponentBuildPhase = "pushing docker image"

	ComponentBuildStatePending   types.ComponentBuildState = "Pending"
	ComponentBuildStateQueued    types.ComponentBuildState = "Queued"
	ComponentBuildStateRunning   types.ComponentBuildState = "Running"
	ComponentBuildStateSucceeded types.ComponentBuildState = "Succeeded"
	ComponentBuildStateFailed    types.ComponentBuildState = "Failed"
)
