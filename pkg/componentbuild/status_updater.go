package componentbuild

import (
	coretypes "github.com/mlab-lattice/core/pkg/types"
)

type Phase string

const (
	PhasePullingGitRepository = "pulling git repository"
	PhaseBuildingDockerImage  = "building docker image"
	PhasePushingDockerImage   = "pushing docker image"
)

type StatusUpdater interface {
	UpdateProgress(coretypes.ComponentBuildID, Phase) error
	UpdateError(buildID coretypes.ComponentBuildID, internal bool, err error) error
}
