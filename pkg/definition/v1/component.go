package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
)

const APIVersion = "v1"

func NewComponentFromJSON(data []byte) (component.Interface, error) {
	var c componentTypeDecoder
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("resource must have valid Type field")
	}

	if c.Type.APIVersion != APIVersion {
		return nil, fmt.Errorf("attempting to create new %v component but APIVersion is %v", APIVersion, c.Type.APIVersion)
	}

	switch c.Type.Type {
	case ComponentTypeJob:
		var j *Job
		if err := json.Unmarshal(data, &j); err != nil {
			return nil, err
		}
		return j, nil

	case ComponentTypeService:
		var s *Service
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, err
		}
		return s, nil

	case ComponentTypeSystem:
		var s *System
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, err
		}
		return s, nil

	default:
		return nil, fmt.Errorf("invalid component type: %v", c.Type.String())
	}
}

type componentTypeDecoder struct {
	Type component.Type `json:"type"`
}
