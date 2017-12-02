package definition

import (
	"errors"
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

const SystemType = "system"

type System struct {
	Meta       block.Metadata `json:"$"`
	Subsystems []Interface    `json:"subsystems"`
}

// Implement json.Unmarshaler
func (s *System) UnmarshalJSON(data []byte) error {
	definition, err := UnmarshalJSON(data)
	if err != nil {
		return err
	}

	if sysDefinition, ok := definition.(*System); ok {
		s.Meta = sysDefinition.Meta
		s.Subsystems = sysDefinition.Subsystems
	}
	return nil
}

// Implement Interface
func (s *System) Metadata() *block.Metadata {
	return &s.Meta
}

// Implement block.Interface
func (s *System) Validate(interface{}) error {
	if s == nil {
		return errors.New("cannot have nil System definition")
	}

	if err := s.Meta.Validate(nil); err != nil {
		return fmt.Errorf("metadata definition error: %v", err)
	}

	if s.Meta.Type != SystemType {
		return fmt.Errorf("expected type %v but got %v", SystemType, s.Meta.Type)
	}

	if len(s.Subsystems) == 0 {
		return errors.New("must have at least one subsystem")
	}

	subsystemNames := map[string]bool{}
	for _, subsystem := range s.Subsystems {
		subsystemName := subsystem.Metadata().Name
		if _, exists := subsystemNames[subsystemName]; exists {
			return fmt.Errorf("multiple subsystems with the name %v", subsystemName)
		}
		subsystemNames[subsystemName] = true

		subsystemDefinitionBlock := subsystem.(block.Interface)
		if err := subsystemDefinitionBlock.Validate(nil); err != nil {
			return fmt.Errorf("subsystem %v definition error: %v", subsystemName, err)
		}
	}

	return nil
}
