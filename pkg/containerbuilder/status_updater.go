package containerbuilder

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type StatusUpdater interface {
	UpdateProgress(v1.ContainerBuildID, v1.SystemID, v1.ContainerBuildPhase) error
	UpdateError(buildID v1.ContainerBuildID, systemID v1.SystemID, internal bool, err error) error
}
