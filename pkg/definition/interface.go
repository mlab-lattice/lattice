package definition

import (
	"encoding/json"
	"fmt"
)

const (
	TypeSystem  = "system"
	TypeService = "service"
)

func NewFromJSON(data []byte) (interface{}, error) {
	var decoded typeDecoder
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}

	if decoded.Type == "" {
		return nil, fmt.Errorf("definition must have a type")
	}

	switch decoded.Type {
	case TypeSystem:
		var system *System
		err := json.Unmarshal(data, &system)
		if err != nil {
			return nil, err
		}

		return system, nil

	case TypeService:
		var service *Service
		err := json.Unmarshal(data, &service)
		if err != nil {
			return nil, err
		}

		return service, nil

	default:
		return nil, fmt.Errorf("unsupported definition type: %v", decoded.Type)
	}
}

type typeDecoder struct {
	Type string `json:"type"`
}
