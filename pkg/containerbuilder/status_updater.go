package containerbuilder

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type StatusUpdater interface {
	UpdateProgress(v1.ComponentBuildID, v1.SystemID, v1.ComponentBuildPhase) error
	UpdateError(buildID v1.ComponentBuildID, systemID v1.SystemID, internal bool, err error) error
}
