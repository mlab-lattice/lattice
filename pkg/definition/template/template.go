package template

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	variableRegex       = regexp.MustCompile(`\$\{([a-zA-Z_$][a-zA-Z_$.0-9]*)\}`)
	singleVariableRegex = regexp.MustCompile(fmt.Sprintf("^%v$", variableRegex.String()))
)

type Template struct {
	Parameters Parameters
	Fields     map[string]interface{}
}

func (t *Template) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if p, ok := m["parameters"]; ok {
		switch params := p.(type) {
		case Parameters:
			t.Parameters = params

		default:
			return fmt.Errorf("invalid parameter type")
		}

		delete(m, "parameters")
	}

	t.Fields = m
	return nil
}

// Evaluate will crawl the Fields map and inject parameters into the values.
// N.B.: Evaluate will mutate the Fields field in place to reduce the number of
// allocations, but returns the result as a convenience.
func (t *Template) Evaluate(assignments map[string]interface{}) (map[string]interface{}, error) {
	assignments, err := t.Parameters.Assign(assignments)
	if err != nil {
		return nil, err
	}

	// no parameters to evaluate
	if len(assignments) == 0 {
		return t.Fields, nil
	}

	// evaluate each of the fields
	for k, v := range t.Fields {
		result, err := evaluateValue(v, assignments)
		if err != nil {
			return nil, err
		}

		t.Fields[k] = result
	}

	return t.Fields, nil
}

func evaluateValue(val interface{}, assignments map[string]interface{}) (interface{}, error) {
	switch v := val.(type) {
	case map[string]interface{}:
		return evaluateMap(v, assignments)

	case []interface{}:
		return evaluateArray(v, assignments)

	case string:
		return evaluateString(v, assignments)

	default:
		return val, nil
	}
}

func evaluateMap(val, assignments map[string]interface{}) (map[string]interface{}, error) {
	for k, v := range val {
		result, err := evaluateValue(v, assignments)
		if err != nil {
			return nil, err
		}

		val[k] = result
	}

	return val, nil
}

func evaluateArray(val []interface{}, assignments map[string]interface{}) ([]interface{}, error) {
	for idx, v := range val {
		result, err := evaluateValue(v, assignments)
		if err != nil {
			return nil, err
		}

		val[idx] = result
	}

	return val, nil
}

func evaluateString(val string, assignments map[string]interface{}) (interface{}, error) {
	// If the string is only a single variable (i.e. "${foo}"), we replace the variable with
	// the true value of the assignment.
	// For example, if foo was set to 3, we would return 3, not "3".
	if singleVariableRegex.MatchString(val) {
		parts := singleVariableRegex.FindStringSubmatch(val)
		variable := parts[1]
		v, ok := assignments[variable]
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

		v, ok := assignments[variableName]
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
