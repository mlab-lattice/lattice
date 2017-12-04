package block

import (
	"errors"
)

type ComponentExec struct {
	Command     []string          `json:"command"`
	Environment map[string]string `json:"environment,omitempty"`
}

// Validate implements Interface
func (e *ComponentExec) Validate(interface{}) error {
	if len(e.Command) == 0 {
		return errors.New("command must have at least one element")
	}

	return nil
}
