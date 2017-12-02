package block

import (
	"errors"
	"fmt"
	"strings"
)

type ComponentHealthCheck struct {
	Http *HttpComponentHealthCheck `json:"http,omitempty"`
	Tcp  *TcpComponentHealthCheck  `json:"tcp,omitempty"`
	Exec *ExecComponentHealthCheck `json:"exec,omitempty"`
}

// Implement Interface
func (hc *ComponentHealthCheck) Validate(information interface{}) error {
	ports := information.(map[string]*ComponentPort)

	if (hc.Http != nil && hc.Tcp != nil) || (hc.Http != nil && hc.Exec != nil) || (hc.Tcp != nil && hc.Exec != nil) {
		return errors.New("only one health_check type may be specified")
	}

	if hc.Http == nil && hc.Tcp == nil && hc.Exec == nil {
		return errors.New("one health_check type must be specified")
	}

	if hc.Http != nil {
		if err := hc.Http.Validate(ports); err != nil {
			return errors.New(fmt.Sprintf("http health_check definition error: %v", err))
		}
	}

	if hc.Tcp != nil {
		if err := hc.Tcp.Validate(ports); err != nil {
			return errors.New(fmt.Sprintf("tcp health_check definition error: %v", err))
		}
	}

	if hc.Exec != nil {
		if err := hc.Exec.Validate(nil); err != nil {
			return errors.New(fmt.Sprintf("exec health_check definition error: %v", err))
		}
	}

	return nil
}

type HttpComponentHealthCheck struct {
	Path    string            `json:"path"`
	Port    string            `json:"port"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Implement Interface
func (hhc *HttpComponentHealthCheck) Validate(information interface{}) error {
	ports := information.(map[string]*ComponentPort)

	port, exists := ports[hhc.Port]
	if !exists {
		return fmt.Errorf("invalid port: %v", hhc.Port)
	}

	if port.Protocol != HttpProtocol {
		return fmt.Errorf("port %s is does not have protocol %v", hhc.Port, HttpProtocol)
	}

	if !strings.HasPrefix(hhc.Path, "/") {
		return errors.New("path must begin with '/'")
	}

	return nil
}

type TcpComponentHealthCheck struct {
	Port string `json:"port"`
}

// Implement Interface
func (thc *TcpComponentHealthCheck) Validate(information interface{}) error {
	ports := information.(map[string]*ComponentPort)

	port, exists := ports[thc.Port]
	if !exists {
		return fmt.Errorf("invalid port: %v", thc.Port)
	}

	// TODO: should we allow TcpComponentHealthCheck on an HTTP port?
	if port.Protocol != TcpProtocol {
		return fmt.Errorf("port %s is does not have protocol %v", thc.Port, TcpProtocol)
	}

	return nil
}

type ExecComponentHealthCheck struct {
	Command []string `json:"command"`
}

// Implement Interface
func (e *ExecComponentHealthCheck) Validate(interface{}) error {
	if len(e.Command) == 0 {
		return errors.New("command must have at least one element")
	}

	return nil
}
