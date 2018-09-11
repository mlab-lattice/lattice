package template

import (
	"fmt"
	"reflect"
	"testing"
)

func TestVariableRegex(t *testing.T) {
	tests := []struct {
		input      string
		numMatches int
	}{
		{"", 0},
		{"foo", 0},
		{"{foo}", 0},
		{"${foo", 0},
		{"$foo", 0},
		{" ${foo}", 1},
		{"${foo} ", 1},
		{"${foo} ${bar}", 2},
		{"${foo}", 1},
	}

	for _, test := range tests {
		matches := variableRegex.FindAllStringSubmatch(test.input, -1)
		if len(matches) != test.numMatches {
			err := fmt.Errorf("expected %v matches for input %v but got %v", test.numMatches, test.input, len(matches))
			t.Error(err)
		}
	}
}

func TestSingleVariableRegex(t *testing.T) {
	tests := []struct {
		input           string
		expectedSuccess bool
	}{
		{"", false},
		{"foo", false},
		{"{foo}", false},
		{"${foo", false},
		{"$foo", false},
		{" ${foo}", false},
		{"${foo} ", false},
		{"${foo} ${bar}", false},
		{"${foo}", true},
	}

	for _, test := range tests {
		match := singleVariableRegex.MatchString(test.input)
		if match != test.expectedSuccess {
			err := fmt.Errorf("expected match for input %v but did not match", test.input)
			if match {
				err = fmt.Errorf("expected no match for input %v but matched", test.input)
			}
			t.Error(err)
		}
	}
}

func TestTemplateEvaluate(t *testing.T) {
	tests := []struct {
		d string
		t Template
		r map[string]interface{}
		a map[string]interface{}
		e bool
	}{
		{
			d: "no parameters no assignments",
			t: Template{
				Parameters: Parameters{},
				Fields: map[string]interface{}{
					"foo": "bar",
					"baz": []interface{}{"buzz"},
					"bat": map[string]interface{}{
						"qux": "hello",
					},
				},
			},
			r: map[string]interface{}{
				"foo": "bar",
				"baz": []interface{}{"buzz"},
				"bat": map[string]interface{}{
					"qux": "hello",
				},
			},
			a: make(map[string]interface{}),
			e: false,
		},
	}

	for _, test := range tests {
		result, err := test.t.Evaluate("", test.a)
		if err != nil {
			if !test.e {
				t.Errorf("got unexpected error evaluating %v: %v", test.d, err)
				continue
			}
		} else {
			if test.e {
				t.Errorf("expected error evaluating %v but got nil", test.d)
				continue
			}
		}

		if !reflect.DeepEqual(result, test.r) {
			t.Errorf("did not get expected result when evaluating %v", test.d)
		}
	}
}
