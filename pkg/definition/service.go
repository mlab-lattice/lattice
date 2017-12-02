package definition

import (
	"errors"
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

const ServiceType = "service"

type Service struct {
	Meta       block.Metadata     `json:"$"`
	Volumes    []*block.Volume    `json:"volumes,omitempty"`
	Components []*block.Component `json:"components"`
	Resources  block.Resources    `json:"resources"`
}

// Implement Interface
func (s *Service) Metadata() *block.Metadata {
	return &s.Meta
}

// Implement block.Interface
func (s *Service) Validate(interface{}) error {
	if s == nil {
		return errors.New("cannot have nil Service definition")
	}

	if err := s.Meta.Validate(nil); err != nil {
		return fmt.Errorf("metadata definition error: %v", err)
	}

	if s.Meta.Type != ServiceType {
		return fmt.Errorf("expected type %v but got %v", ServiceType, s.Meta.Type)
	}

	volumes := map[string]*block.Volume{}
	for _, volume := range s.Volumes {
		if err := volume.Validate(nil); err != nil {
			return fmt.Errorf("volume %v definition error: %v", volume.Name, err)
		}
		volumes[volume.Name] = volume
	}

	if len(s.Components) == 0 {
		return errors.New("must specify at least one component")
	}

	for componentName, component := range s.Components {
		if err := component.Validate(volumes); err != nil {
			return fmt.Errorf("component %v definition error: %v", componentName, err)
		}
	}

	if err := s.Resources.Validate(nil); err != nil {
		return fmt.Errorf("resources definition error: %v", err)
	}

	return nil
}
