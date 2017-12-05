package block

import (
	"errors"
)

type Metadata struct {
	Name        string                       `json:"name"`
	Type        string                       `json:"type"`
	Description string                       `json:"description"`
	Parameters  map[string]MetadataParameter `json:"parameters,omitempty"`
}

// Validate implements Interface
func (m *Metadata) Validate(interface{}) error {
	if m.Name == "" {
		return errors.New("name is required")
	}

	if m.Type == "" {
		return errors.New("type is required")
	}

	return nil
}

type MetadataParameter struct {
	// TODO: add type
	// TODO: add default
	Description string `json:"description"`
}

// Validate implements Interface
func (m *MetadataParameter) Validate(interface{}) error {
	// TODO: add parameter validation
	return nil
}
