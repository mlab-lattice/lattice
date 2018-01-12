package definition

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

type Service interface {
	Interface
	Volumes() []*block.Volume
	Components() []*block.Component
	Resources() block.Resources
}

type ServiceValidator interface {
	Validate(Service) error
}

func NewServiceFromJSON(data []byte) (Service, error) {
	var decoded serviceDecoder
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}

	if decoded.Type != TypeService {
		return nil, fmt.Errorf("service type must be %v", TypeService)
	}

	s := &service{
		name:       decoded.Name,
		volumes:    decoded.Volumes,
		components: decoded.Components,
		resources:  decoded.Resources,
	}
	return s, nil
}

type serviceDecoder struct {
	Type       string             `json:"type"`
	Name       string             `json:"name"`
	Volumes    []*block.Volume    `json:"volumes"`
	Components []*block.Component `json:"components"`
	Resources  block.Resources    `json:"resources"`
}

type service struct {
	name       string
	volumes    []*block.Volume
	components []*block.Component
	resources  block.Resources
}

func (s *service) Type() string {
	return TypeService
}

func (s *service) Name() string {
	return s.name
}

func (s *service) Volumes() []*block.Volume {
	return s.volumes
}

func (s *service) Components() []*block.Component {
	return s.components
}

func (s *service) Resources() block.Resources {
	return s.resources
}
