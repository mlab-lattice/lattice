package block

import (
	"encoding/json"
	"errors"
)

type ComponentExec struct {
	Command     []string          `json:"command"`
	Environment map[string]string `json:"environment,omitempty"`
}

type Environment map[string]EnvironmentVariable

type EnvironmentVariable struct {
	Value  *string
	Secret *Secret
}

func (ev *EnvironmentVariable) UnmarshalJSON(data []byte) error {
	eve := &environmentVariableEncoder{}
	if err := json.Unmarshal(data, &eve); err != nil {
		return err
	}

	if eve.Value != nil {
		ev.Value = eve.Value
	}

	if eve.Secret != nil {
		ev.Secret = eve.Secret
	}

	return nil
}

type environmentVariableEncoder struct {
	Value  *string
	Secret *Secret
}

func (eve *environmentVariableEncoder) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &eve.Value)
	if err != nil {
		if _, ok := err.(*json.UnmarshalTypeError); !ok {
			return err
		}

		err = json.Unmarshal(data, &eve.Secret)
	}

	return err
}

// Validate implements Interface
func (e *ComponentExec) Validate(interface{}) error {
	if len(e.Command) == 0 {
		return errors.New("command must have at least one element")
	}

	return nil
}
