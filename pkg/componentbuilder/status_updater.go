package componentbuilder

import (
	"github.com/mlab-lattice/system/pkg/types"
)

type StatusUpdater interface {
	UpdateProgress(types.ComponentBuildID, types.ComponentBuildPhase) error
	UpdateError(buildID types.ComponentBuildID, internal bool, err error) error
}
