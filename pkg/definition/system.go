package definition

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/block"
)

type System interface {
	Interface
	NodePools() map[string]block.NodePool
	Subsystems() []Interface
}

type SystemValidator interface {
	Validate(System) error
}

func NewSystemFromJSON(data []byte) (System, error) {
	var decoded systemEncoder
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}

	if decoded.Type != TypeSystem {
		return nil, fmt.Errorf("system type must be %v", TypeSystem)
	}

	var subsystems []Interface
	for _, subsystemJSON := range decoded.Subsystems {
		subsystem, err := NewFromJSON(subsystemJSON)
		if err != nil {
			return nil, err
		}

		subsystems = append(subsystems, subsystem)
	}

	s := &system{
		name:        decoded.Name,
		description: decoded.Description,

		nodePools:  decoded.NodePools,
		subsystems: subsystems,
	}
	return s, nil
}

type systemEncoder struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`

	NodePools  map[string]block.NodePool `json:"node_pools"`
	Subsystems []json.RawMessage         `json:"subsystems"`
}

type system struct {
	name        string
	description string

	nodePools  map[string]block.NodePool
	subsystems []Interface
}

func (s *system) Type() string {
	return TypeSystem
}

func (s *system) Name() string {
	return s.name
}

func (s *system) Description() string {
	return s.description
}

func (s *system) NodePools() map[string]block.NodePool {
	return s.nodePools
}

func (s *system) Subsystems() []Interface {
	return s.subsystems
}

func (s *system) MarshalJSON() ([]byte, error) {
	var subsystems []json.RawMessage
	for _, subsystem := range s.subsystems {
		subsystemJSON, err := json.Marshal(subsystem)
		if err != nil {
			return nil, err
		}

		subsystems = append(subsystems, subsystemJSON)
	}

	encoder := systemEncoder{
		Type:        TypeSystem,
		Name:        s.name,
		Description: s.description,
		NodePools:   s.nodePools,
		Subsystems:  subsystems,
	}

	return json.Marshal(&encoder)
}
