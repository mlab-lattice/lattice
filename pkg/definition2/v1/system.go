package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition2/resource"
)

const ResourceTypeSystem = "system"

var SystemType = resource.Type{
	APIVersion: APIVersion,
	Type:       ResourceTypeSystem,
}

type System struct {
	Description string

	Subsystems map[string]resource.Interface
}

func (s *System) Type() resource.Type {
	return SystemType
}

func (s *System) MarshalJSON() ([]byte, error) {
	e := systemEncoder{
		Type:        SystemType,
		Description: s.Description,

		Subsystems: s.Subsystems,
	}
	return json.Marshal(&e)
}

func (s *System) UnmarshalJSON(data []byte) error {
	var e *systemEncoder
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	if e.Type.APIVersion != APIVersion {
		return fmt.Errorf("expected api version %v but got %v", APIVersion, e.Type.APIVersion)
	}

	if e.Type.Type != ResourceTypeSystem {
		return fmt.Errorf("expected resource type %v but got %v", ResourceTypeSystem, e.Type.Type)
	}

	system := &System{
		Description: e.Description,

		Subsystems: e.Subsystems,
	}
	*s = *system
	return nil
}

type systemEncoder struct {
	Type        resource.Type `json:"type"`
	Description string        `json:"description,omitempty"`

	Subsystems map[string]resource.Interface `json:"subsystems"`
}
