package block

import (
	"fmt"
	"strings"
)

type ComponentHealthCheck struct {
	HTTP *HTTPComponentHealthCheck `json:"http,omitempty"`
	TCP  *TCPComponentHealthCheck  `json:"tcp,omitempty"`
	Exec *ExecComponentHealthCheck `json:"exec,omitempty"`
}

// Validate implements Interface
func (hc *ComponentHealthCheck) Validate(information interface{}) error {
	ports := information.(map[string]*ComponentPort)

	if (hc.HTTP != nil && hc.TCP != nil) || (hc.HTTP != nil && hc.Exec != nil) || (hc.TCP != nil && hc.Exec != nil) {
		return fmt.Errorf("only one health_check type may be specified")
	}

	if hc.HTTP == nil && hc.TCP == nil && hc.Exec == nil {
		return fmt.Errorf("one health_check type must be specified")
	}

	if hc.HTTP != nil {
		if err := hc.HTTP.Validate(ports); err != nil {
			return fmt.Errorf("http health_check definition error: %v", err)
		}
	}

	if hc.TCP != nil {
		if err := hc.TCP.Validate(ports); err != nil {
			return fmt.Errorf("tcp health_check definition error: %v", err)
		}
	}

	if hc.Exec != nil {
		if err := hc.Exec.Validate(nil); err != nil {
			return fmt.Errorf("exec health_check definition error: %v", err)
		}
	}

	return nil
}

type HTTPComponentHealthCheck struct {
	Path    string            `json:"path"`
	Port    string            `json:"port"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Validate implements Interface
func (hhc *HTTPComponentHealthCheck) Validate(information interface{}) error {
	ports := information.(map[string]*ComponentPort)

	port, exists := ports[hhc.Port]
	if !exists {
		return fmt.Errorf("invalid port: %v", hhc.Port)
	}

	if port.Protocol != ProtocolHTTP {
		return fmt.Errorf("port %s is does not have protocol %v", hhc.Port, ProtocolHTTP)
	}

	if !strings.HasPrefix(hhc.Path, "/") {
		return fmt.Errorf("path must begin with '/'")
	}

	return nil
}

type TCPComponentHealthCheck struct {
	Port string `json:"port"`
}

// Validate implements Interface
func (thc *TCPComponentHealthCheck) Validate(information interface{}) error {
	ports := information.(map[string]*ComponentPort)

	port, exists := ports[thc.Port]
	if !exists {
		return fmt.Errorf("invalid port: %v", thc.Port)
	}

	// TODO: should we allow TCPComponentHealthCheck on an HTTP port?
	if port.Protocol != ProtocolTCP {
		return fmt.Errorf("port %s is does not have protocol %v", thc.Port, ProtocolTCP)
	}

	return nil
}

type ExecComponentHealthCheck struct {
	Command []string `json:"command"`
}

// Validate implements Interface
func (e *ExecComponentHealthCheck) Validate(interface{}) error {
	if len(e.Command) == 0 {
		return fmt.Errorf("command must have at least one element")
	}

	return nil
}
