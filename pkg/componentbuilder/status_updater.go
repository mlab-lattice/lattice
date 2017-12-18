package componentbuilder

import (
	"github.com/mlab-lattice/system/pkg/types"
)

type StatusUpdater interface {
	UpdateProgress(types.ComponentBuildID, types.SystemID, types.ComponentBuildPhase) error
	UpdateError(buildID types.ComponentBuildID, systemID types.SystemID, internal bool, err error) error
}
