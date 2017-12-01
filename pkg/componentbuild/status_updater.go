package componentbuild

import (
	coretypes "github.com/mlab-lattice/core/pkg/types"
)

type StatusUpdater interface {
	UpdateProgress(coretypes.ComponentBuildID, coretypes.ComponentBuildPhase) error
	UpdateError(buildID coretypes.ComponentBuildID, internal bool, err error) error
}
