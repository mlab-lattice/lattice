package definition

import (
	"encoding/json"
	"fmt"
)

const (
	TypeSystem  = "system"
	TypeService = "service"
)

type Interface interface {
	Type() string
	Name() string
}

type Validator interface {
	Validate(Interface) error
}

func NewFromJSON(data []byte) (Interface, error) {
	var decoded typeDecoder
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}

	if decoded.Type == "" {
		return nil, fmt.Errorf("definition must have a type")
	}

	var definition Interface
	switch decoded.Type {
	case TypeSystem:
		system, err := NewSystemFromJSON(data)
		if err != nil {
			return nil, err
		}

		definition = system.(Interface)

	case TypeService:
		service, err := NewServiceFromJSON(data)
		if err != nil {
			return nil, err
		}

		definition = service.(Interface)

	default:
		return nil, fmt.Errorf("unsupported definition type: %v", decoded.Type)
	}

	return definition, nil
}

type typeDecoder struct {
	Type      string `json:"type"`
	Remainder json.RawMessage
}
