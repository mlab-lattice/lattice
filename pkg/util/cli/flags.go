package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

type EmbeddedFlag struct {
	Expected map[string]EmbeddedFlagValue
	Target   interface{}

	expected      map[string]struct{}
	required      map[string]struct{}
	arrayVars     map[string][]string
	encodingNames map[string]string
}

func (ef *EmbeddedFlag) init() {
	ef.expected = map[string]struct{}{}
	ef.required = map[string]struct{}{}
	ef.arrayVars = map[string][]string{}
	ef.encodingNames = map[string]string{}

	for key, value := range ef.Expected {
		ef.expected[key] = struct{}{}

		if value.Required {
			ef.required[key] = struct{}{}
		}

		ef.encodingNames[key] = key
		if value.EncodingName != "" {
			ef.encodingNames[key] = value.EncodingName
		}

	}
}

func (ef *EmbeddedFlag) Parse(values []string) error {
	ef.init()

	result := map[string]interface{}{}
	for _, value := range values {
		key, value, err := parseEmbeddedFlagValue(value)
		if err != nil {
			return err
		}

		flag, ok := ef.Expected[key]
		if !ok {
			return fmt.Errorf("unexpected key %v", key)
		}

		// set a marker in the result map so that we know that we have a value for this key,
		// add the value to our list of values for this key, and keep on going (we'll parse the whole slice later)
		if flag.Array {
			result[key] = struct{}{}
			ef.arrayVars[key] = append(ef.arrayVars[key], value)
			continue
		}

		parsed := interface{}(value)
		if flag.ValueParser != nil {
			parsed, err = flag.ValueParser(value)
			if err != nil {
				return fmt.Errorf("error parsing %v: %v", key, err)
			}
		}

		result[key] = parsed
	}

	for requiredKey := range ef.required {
		if _, ok := result[requiredKey]; !ok {
			return fmt.Errorf("missing required key %v", requiredKey)
		}
	}

	for key, arrayVal := range ef.arrayVars {
		flag := ef.Expected[key]
		parsed := interface{}(arrayVal)
		var err error
		if flag.ArrayValueParser != nil {
			parsed, err = flag.ArrayValueParser(arrayVal)
			if err != nil {
				return fmt.Errorf("error parsing %v: %v", key, err)
			}
		}

		result[key] = parsed
	}

	for expectedKey := range ef.expected {
		if _, ok := result[expectedKey]; !ok {
			result[expectedKey] = ef.Expected[expectedKey].Default
		}

		if val, ok := result[expectedKey]; ok {
			delete(result, expectedKey)
			result[ef.encodingNames[expectedKey]] = val
		}
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, ef.Target)
}

type EmbeddedFlagValue struct {
	ValueParser      func(string) (interface{}, error)
	ArrayValueParser func([]string) (interface{}, error)
	Required         bool
	Default          interface{}
	Array            bool
	EncodingName     string
}

func parseEmbeddedFlagValue(value string) (string, string, error) {
	split := strings.Split(value, "=")
	if len(split) < 2 {
		return "", "", fmt.Errorf("expected form key=value, but got %v", value)
	}

	return split[0], strings.Join(split[1:], "="), nil
}
