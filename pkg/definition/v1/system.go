package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/resource"
)

const ResourceTypeSystem = "system"

var SystemType = resource.Type{
	APIVersion: APIVersion,
	Type:       ResourceTypeSystem,
}

type System struct {
	Description string

	Resources map[string]resource.Interface
}

func (s *System) Type() resource.Type {
	return SystemType
}

func (s *System) MarshalJSON() ([]byte, error) {
	e := systemEncoder{
		Type:        SystemType,
		Description: s.Description,

		Resources: s.Resources,
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

	if e.Type.Type != ResourceTypeSystem {
		return fmt.Errorf("expected resource type %v but got %v", ResourceTypeSystem, e.Type.Type)
	}

	resources := make(map[string]resource.Interface)
	for n, d := range e.Resources {
		res, err := NewResource(d)
		if err != nil {
			return err
		}

		resources[n] = res
	}

	system := &System{
		Description: e.Description,

		Resources: resources,
	}
	*s = *system
	return nil
}

type systemEncoder struct {
	Type        resource.Type `json:"type"`
	Description string        `json:"description,omitempty"`

	Resources map[string]resource.Interface `json:"resources"`
}

type systemDecoder struct {
	Type        resource.Type `json:"type"`
	Description string        `json:"description,omitempty"`

	Resources map[string]json.RawMessage `json:"resources"`
}
