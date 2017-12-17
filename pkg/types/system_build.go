package types

import (
	// TODO: feels a little weird to have to import this here. should type definitions under pkg/system be moved into pkg/types?
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type SystemBuildID string
type SystemBuildState string

const (
	SystemBuildStatePending   SystemBuildState = "Pending"
	SystemBuildStateRunning   SystemBuildState = "Running"
	SystemBuildStateSucceeded SystemBuildState = "Succeeded"
	SystemBuildStateFailed    SystemBuildState = "Failed"
)

type SystemBuild struct {
	ID    SystemBuildID    `json:"id"`
	State SystemBuildState `json:"state"`

	Version SystemVersion `json:"version"`
	// Services maps service paths (e.g. /foo/bar/buzz) to the
	// ServiceBuild for that service in the SystemBuild.
	Services map[tree.NodePath]ServiceBuild `json:"serviceBuilds"`
}
