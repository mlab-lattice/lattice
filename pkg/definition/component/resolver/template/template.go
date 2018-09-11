package template

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

const parametersField = "$parameters"

var (
	variableRegex       = regexp.MustCompile(`\$\{([a-zA-Z_$][a-zA-Z_$.0-9]*)\}`)
	singleVariableRegex = regexp.MustCompile(fmt.Sprintf("^%v$", variableRegex.String()))
)

type Template struct {
	Parameters Parameters
	Fields     map[string]interface{}
}

// UnmarshalJSON implements the json.Unmarshaller interface.
// The JSON should be an object. UnmarshalJSON will see if
// the object has a parameters field, and if it does try to
// Unmarshal it into a Parameters struct. If it succeeds it will
// use this struct as the Template's Parameters field. It will
// then use the rest of the fields of the object as the Template's
// Fields field.
func (t *Template) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	params := make(Parameters)
	if p, ok := m[parametersField]; ok && p != nil {
		if _, ok := p.(map[string]interface{}); !ok {
			return fmt.Errorf("invalid parameter type, expected object")
		}

		paramBytes, err := json.Marshal(&p)
		if err != nil {
			return fmt.Errorf("error marshalling parameters: %v", err)
		}

		if err := json.Unmarshal(paramBytes, &params); err != nil {
			return err
		}
	}

	delete(m, parametersField)
	t.Parameters = params
	t.Fields = m

	return nil
}

func (t *Template) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	for k, v := range t.Fields {
		m[k] = v
	}

	m[parametersField] = t.Parameters
	return json.Marshal(&m)
}

// Evaluate will crawl the Fields map and inject parameters into the values.
func (t *Template) Evaluate(path tree.Path, bindings map[string]interface{}) (map[string]interface{}, error) {
	bindings, err := t.Parameters.Bind(path, bindings)
	if err != nil {
		return nil, err
	}

	// evaluate each of the fields
	for k, v := range t.Fields {
		result, err := evaluateValue(path, v, bindings)
		if err != nil {
			return nil, err
		}

		t.Fields[k] = result
	}

	return t.Fields, nil
}

func evaluateValue(path tree.Path, val interface{}, bindings map[string]interface{}) (interface{}, error) {
	switch v := val.(type) {
	case map[string]interface{}:
		return evaluateMap(path, v, bindings)

	case []interface{}:
		return evaluateArray(path, v, bindings)

	case string:
		return evaluateString(path, v, bindings)

	default:
		return val, nil
	}
}

func evaluateMap(path tree.Path, val, bindings map[string]interface{}) (interface{}, error) {
	for k, v := range val {
		// check to see if the map is a $secret
		// TODO(kevindrosendahl): this is too tightly coupled, think of how to refactor
		// TODO(kevindrosendahl): should check to make sure that $secret is the only key in the map
		if k == SecretParameterLVal {
			// Ensure the $secret key is a string
			// TODO(kevindrosendahl): validate character set here?
			name, ok := v.(string)
			if !ok {
				return nil, &ParameterTypeError{
					Expected: string(ParameterTypeSecret),
					Actual:   reflect.TypeOf(v).String(),
				}
			}

			secretRefPath, err := tree.NewPathSubcomponentFromParts(path, name)
			if err != nil {
				return nil, err
			}

			secretRef := &definitionv1.SecretRef{
				Value: secretRefPath,
			}
			return secretRef, nil
		}

		result, err := evaluateValue(path, v, bindings)
		if err != nil {
			return nil, err
		}

		val[k] = result
	}

	return val, nil
}

func evaluateArray(path tree.Path, val []interface{}, bindings map[string]interface{}) (interface{}, error) {
	for idx, v := range val {
		result, err := evaluateValue(path, v, bindings)
		if err != nil {
			return nil, err
		}

		val[idx] = result
	}

	return val, nil
}

func evaluateString(path tree.Path, val string, bindings map[string]interface{}) (interface{}, error) {
	// If the string is only a single variable (i.e. "${foo}"), we replace the variable with
	// the true value of the assignment.
	// For example, if foo was set to 3, we would return 3, not "3".
	if singleVariableRegex.MatchString(val) {
		parts := singleVariableRegex.FindStringSubmatch(val)
		variable := parts[1]
		v, ok := bindings[variable]
		if !ok {
			return nil, fmt.Errorf("invalid template variable %v", variable)
		}

		return v, nil
	}

	// If the string is not a single variable, find all of the variables in the string
	// and replace them with the string representation of the value.
	// For now we take the string representation of the value to mean the JSON encoding.
	variables := variableRegex.FindAllStringSubmatch(val, -1)
	for _, variableParts := range variables {
		variableString := variableParts[0]
		variableName := variableParts[1]

		v, ok := bindings[variableName]
		if !ok {
			return nil, fmt.Errorf("invalid template variable %v", variableName)
		}

		e, err := json.Marshal(&v)
		if err != nil {
			return nil, err
		}

		val = strings.Replace(val, variableString, string(e), -1)
	}

	return val, nil
}
