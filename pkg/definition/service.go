package definition

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/block"
)

// Note: if you change anything here, update serviceEncoder as well
type Service struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Components []block.Component       `json:"components"`
	Resources  block.Resources         `json:"resources"`
	Secrets    map[string]block.Secret `json:"secrets"`
}

func (s *Service) UnmarshalJSON(data []byte) error {
	var e *serviceEncoder
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	if e.Type != TypeService {
		return fmt.Errorf("expected type to be %v but got %v", TypeService, e.Type)
	}

	Service := &Service{
		Type:        e.Type,
		Name:        e.Name,
		Description: e.Description,

		Components: e.Components,
		Resources:  e.Resources,
		Secrets:    e.Secrets,
	}
	*s = *Service
	return nil
}

type serviceEncoder struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Components []block.Component       `json:"components"`
	Resources  block.Resources         `json:"resources"`
	Secrets    map[string]block.Secret `json:"secrets"`
}
