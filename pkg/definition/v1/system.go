package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
)

const ComponentTypeSystem = "system"

var SystemType = component.Type{
	APIVersion: APIVersion,
	Type:       ComponentTypeSystem,
}

type System struct {
	Description string

	Components map[string]component.Interface
}

func (s *System) Type() component.Type {
	return SystemType
}

func (s *System) MarshalJSON() ([]byte, error) {
	e := systemEncoder{
		Type:        SystemType,
		Description: s.Description,

		Components: s.Components,
	}
	return json.Marshal(&e)
}

func (s *System) UnmarshalJSON(data []byte) error {
	var e *systemDecoder
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	if e.Type.APIVersion != APIVersion {
		return fmt.Errorf("expected api version %v but got %v", APIVersion, e.Type.APIVersion)
	}

	if e.Type.Type != ComponentTypeSystem {
		return fmt.Errorf("expected resource type %v but got %v", ComponentTypeSystem, e.Type.Type)
	}

	resources := make(map[string]component.Interface)
	for n, d := range e.Components {
		res, err := NewComponentFromJSON(d)
		if err != nil {
			return err
		}

		resources[n] = res
	}

	system := &System{
		Description: e.Description,

		Components: resources,
	}
	*s = *system
	return nil
}

type systemEncoder struct {
	Type        component.Type `json:"type"`
	Description string         `json:"description,omitempty"`

	Components map[string]component.Interface `json:"components"`
}

type systemDecoder struct {
	Type        component.Type `json:"type"`
	Description string         `json:"description,omitempty"`

	Components map[string]json.RawMessage `json:"components"`
}
