package definition

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/block"
)

// Note: if you change anything here, update systemEncoder as well
type System struct {
	Type        string
	Name        string
	Description string

	NodePools  map[string]block.NodePool
	Subsystems []interface{}
}

func (s *System) UnmarshalJSON(data []byte) error {
	var e *systemEncoder
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	if e.Type != TypeSystem {
		return fmt.Errorf("expected type to be %v but got %v", TypeSystem, e.Type)
	}

	var subsystems []interface{}
	for _, subsystemJSON := range e.Subsystems {
		subsystem, err := NewFromJSON(subsystemJSON)
		if err != nil {
			return err
		}

		subsystems = append(subsystems, subsystem)
	}

	system := &System{
		Type:        e.Type,
		Name:        e.Name,
		Description: e.Description,

		NodePools:  e.NodePools,
		Subsystems: subsystems,
	}
	*s = *system
	return nil
}

type systemEncoder struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`

	NodePools  map[string]block.NodePool `json:"node_pools"`
	Subsystems []json.RawMessage         `json:"subsystems"`
}
