package definition

import (
	"encoding/json"
	"fmt"
)

type System interface {
	Interface
	Subsystems() []Interface
}

type SystemValidator interface {
	Validate(System) error
}

func NewSystemFromJSON(data []byte) (System, error) {
	var decoded systemDecoder
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
		name:       decoded.Name,
		subsystems: subsystems,
	}
	return s, nil
}

type systemDecoder struct {
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Subsystems []json.RawMessage `json:"subsystems"`
}

type system struct {
	name       string
	subsystems []Interface
}

func (s *system) Type() string {
	return TypeSystem
}

func (s *system) Name() string {
	return s.name
}

func (s *system) Subsystems() []Interface {
	return s.subsystems
}
