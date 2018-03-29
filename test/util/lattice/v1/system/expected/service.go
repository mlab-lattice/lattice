package expected

import (
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type Service struct {
	Path               tree.NodePath
	ValidStates        []v1.ServiceState
	DesiredState       v1.ServiceState
	UpdatedInstances   Int32Range
	StaleInstances     Int32Range
	ValidPublicPorts   []int32
	DesiredPublicPorts []int32
}
