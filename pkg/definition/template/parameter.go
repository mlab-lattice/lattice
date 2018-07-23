package template

import (
	"fmt"
	"reflect"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type ParameterType string

const (
	ParameterTypeBool   ParameterType = "bool"
	ParameterTypeNumber ParameterType = "number"
	ParameterTypeString ParameterType = "string"
	ParameterTypeArray  ParameterType = "array"
	ParameterTypeObject ParameterType = "object"
	ParameterTypeSecret ParameterType = "secret"

	secretParameterLVal = "$secret"
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

	case *definitionv1.SecretRef:
		if d.Type != ParameterTypeSecret {
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

// Bind will take in a set of bindings for parameters, type check them, and properly set defaults if necessary.
// It will then return the checked and defaulted set of parameter bindings.
func (p Parameters) Bind(path tree.NodePath, bindings map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for k, p := range p {
		// Check to see if an binding was supplied.
		// If one was not but a default value was provided, use the default instead.
		var a interface{}
		var ok bool
		a, ok = bindings[k]
		if !ok {
			if p.Default == nil {
				return nil, fmt.Errorf("missing assignment to parameter %v", k)
			}

			// If the parameter is a secret, its default value should contain something like:
			// { "$secret": "name" }.
			// Convert this into a SecretReference: { "$secret_ref": "/path:name" }
			def := p.Default
			if p.Type == ParameterTypeSecret {
				// Ensure the default value is a map
				secret, ok := def.(map[string]interface{})
				if !ok {
					return nil, &ParameterTypeError{
						Expected: string(ParameterTypeSecret),
						Actual:   reflect.TypeOf(def).String(),
					}
				}

				// Ensure the map has a $secret key
				nameVal, ok := secret[secretParameterLVal]
				if !ok {
					return nil, &ParameterTypeError{
						Expected: string(ParameterTypeSecret),
						Actual:   reflect.TypeOf(def).String(),
					}
				}

				// Ensure the $secret key is a string
				// TODO(kevinrosendahl): validate character set here?
				name, ok := nameVal.(string)
				if !ok {
					return nil, &ParameterTypeError{
						Expected: string(ParameterTypeSecret),
						Actual:   reflect.TypeOf(def).String(),
					}
				}

				secretRefPath, err := tree.NewNodePathSubcomponentFromParts(path, name)
				if err != nil {
					return nil, err
				}

				def = &definitionv1.SecretRef{
					Value: secretRefPath,
				}
			}

			a = def
		}

		if err := p.Validate(a); err != nil {
			return nil, err
		}

		result[k] = a
	}

	// TODO(kevinrosendahl): we may want to validate that there are no extra bindings
	return result, nil
}
