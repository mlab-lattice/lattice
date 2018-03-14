package block

import (
	"encoding/json"
	"errors"
)

type ComponentExec struct {
	Command     []string    `json:"command"`
	Environment Environment `json:"environment,omitempty"`
}

type Environment map[string]*EnvironmentVariable

type EnvironmentVariable struct {
	Value  *string
	Secret *SecretValue
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

func (ev *EnvironmentVariable) MarshalJSON() ([]byte, error) {
	eve := &environmentVariableEncoder{
		Value:  ev.Value,
		Secret: ev.Secret,
	}
	return json.Marshal(&eve)
}

type environmentVariableEncoder struct {
	Value  *string
	Secret *SecretValue
}

func (eve *environmentVariableEncoder) UnmarshalJSON(data []byte) error {
	originalValue := eve.Value
	// First, try to unmarshal it into Name to see if the value
	// is just a string (aka the name of a secret)
	err := json.Unmarshal(data, &eve.Value)
	if err != nil {
		// If Unmarshalling failed due to a type error, that means that
		// we were trying to unmarshal something that was not a string.
		// So we handle this error and keep going.
		if _, ok := err.(*json.UnmarshalTypeError); !ok {
			return err
		}

		// A failed Unmarshal can leave some weird data leftover, so
		// if it failed, reset sve.Name to whatever it was before
		// the attempt.
		eve.Value = originalValue

		// Then, try to Unmarshal the value into the reference field.
		err = json.Unmarshal(data, &eve.Secret)
	}

	return err
}

func (eve *environmentVariableEncoder) MarshalJSON() ([]byte, error) {
	if eve.Value != nil {
		return json.Marshal(*eve.Value)
	}

	return json.Marshal(eve.Secret)
}

// Validate implements Interface
func (e *ComponentExec) Validate(interface{}) error {
	if len(e.Command) == 0 {
		return errors.New("command must have at least one element")
	}

	return nil
}
