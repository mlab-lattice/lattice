package block

import (
	"errors"
	"fmt"
)

type ComponentPort struct {
	Name           string          `json:"name"`
	Port           int32           `json:"port"`
	Protocol       string          `json:"protocol"`
	ExternalAccess *ExternalAccess `json:"external_access,omitempty"`
}

const (
	MinPortNumber = 1
	MaxPortNumber = 65535

	ProtocolHTTP = "HTTP"
	ProtocolTCP  = "TCP"
)

var ValidPortProtocols = map[string]bool{
	ProtocolTCP:  true,
	ProtocolHTTP: true,
}

// Validate implements Interface
func (p *ComponentPort) Validate(interface{}) error {
	if p.Name == "" {
		return errors.New("name is required")
	}

	if MinPortNumber > p.Port || MaxPortNumber < p.Port {
		return fmt.Errorf("invalid port %v", p.Port)
	}

	if _, exists := ValidPortProtocols[p.Protocol]; !exists {
		return fmt.Errorf("invalid protocol %v", p.Protocol)
	}

	if err := p.ExternalAccess.Validate(nil); err != nil {
		return fmt.Errorf("external_access definition error: %v", err)
	}

	return nil
}

// TODO: add peering
type ExternalAccess struct {
	Public bool `json:"public"`
}

// Validate implements Interface
func (ea *ExternalAccess) Validate(interface{}) error {
	return nil
}
