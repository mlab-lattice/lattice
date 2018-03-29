package expected

import "github.com/mlab-lattice/system/pkg/api/v1"

type System struct {
	ID              v1.SystemID
	ValidStates     []v1.SystemState
	DesiredState    v1.SystemState
	DefinitionURL   string
	ValidServices   []Service
	DesiredServices []Service
}
