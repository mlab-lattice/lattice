package componentbuilder

import (
	"github.com/mlab-lattice/system/pkg/types"
)

type StatusUpdater interface {
	UpdateProgress(types.ComponentBuildID, types.LatticeNamespace, types.ComponentBuildPhase) error
	UpdateError(buildID types.ComponentBuildID, namespace types.LatticeNamespace, internal bool, err error) error
}
