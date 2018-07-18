package template

import (
	"fmt"
	"reflect"
)

type ParameterType string

const (
	ParameterTypeBool   ParameterType = "bool"
	ParameterTypeNumber ParameterType = "number"
	ParameterTypeString ParameterType = "string"
	ParameterTypeArray  ParameterType = "array"
	ParameterTypeObject ParameterType = "object"
	//ParameterTypeSecret ParameterType = "secret"
)

type ParameterTypeError struct {
	Expected string
	Actual   string
}

func (e *ParameterTypeError) Error() string {
	return fmt.Sprintf("expected a %v but got a %v", e.Expected, e.Actual)
}

type Parameter struct {
	Type    ParameterType `json:"type"`
	Default interface{}   `json:"default"`
}

func (d Parameter) Validate(assignment interface{}) error {
	failed := false

	// https://golang.org/pkg/encoding/json/#Unmarshal outlines what golang types
	// different json value types will be encoded into
	switch assignment.(type) {
	case bool:
		if d.Type != ParameterTypeBool {
			failed = true
		}

	case float64:
		if d.Type != ParameterTypeNumber {
			failed = true
		}

	case string:
		if d.Type != ParameterTypeString {
			failed = true
		}

	case []interface{}:
		if d.Type != ParameterTypeArray {
			failed = true
		}

	case map[string]interface{}:
		if d.Type != ParameterTypeObject {
			failed = true
		}

	default:
		failed = true
	}

	if !failed {
		return nil
	}

	return &ParameterTypeError{
		Expected: string(d.Type),
		Actual:   reflect.TypeOf(assignment).String(),
	}
}

type Parameters map[string]Parameter

func (p Parameters) Assign(assignments map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for k, p := range p {
		// Check to see if an assignment was supplied.
		// If one was not but a default value was provided, use the default instead.
		var a interface{}
		var ok bool
		a, ok = assignments[k]
		if !ok {
			if p.Default == nil {
				return nil, fmt.Errorf("missing assignment to parameter %v", k)
			}

			a = p.Default
		}

		if err := p.Validate(a); err != nil {
			return nil, err
		}

		result[k] = a
	}

	// TODO(kevinrosendahl): we may want to validate that there are no extra assignments
	return result, nil
}
