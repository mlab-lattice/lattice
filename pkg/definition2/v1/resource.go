package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition2/resource"
)

const APIVersion = "v1"

func NewResource(data []byte) (resource.Interface, error) {
	var r resourceTypeDecoder
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("resource must have valid Type field")
	}

	if r.Type.APIVersion != APIVersion {
		return nil, fmt.Errorf("attempting to create new %v resource but APIVersion is %v", APIVersion, r.Type.APIVersion)
	}

	switch r.Type.Type {
	case ResourceTypeJob:
		var j *Job
		if err := json.Unmarshal(data, &j); err != nil {
			return nil, err
		}
		return j, nil

	case ResourceTypeService:
		var s *Service
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, err
		}
		return s, nil

	case ResourceTypeSystem:
		var s *System
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, err
		}
		return s, nil

	default:
		return nil, fmt.Errorf("invalid resource type: %v", r.Type.String())
	}
}

type resourceTypeDecoder struct {
	Type resource.Type `json:"type"`
}
