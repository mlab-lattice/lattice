package definition

import (
	"fmt"
	"testing"
)

type fromJSONTest struct {
	description     string
	bytes           []byte
	additionalTests []additionalFromJSONTest
}

type additionalFromJSONTest struct {
	description string
	test        func(interface{}) error
}

func expectFromJSONSuccesses(t *testing.T, tests []fromJSONTest, testFunc func([]byte) (interface{}, error)) {
	for _, test := range tests {
		expectFromJSONSuccess(t, &test, testFunc)
	}
}

func expectFromJSONSuccess(t *testing.T, test *fromJSONTest, testFunc func([]byte) (interface{}, error)) {
	result, err := testFunc(test.bytes)
	if err != nil {
		t.Errorf("expected \"%v\" to succeed but it did not: %v", test.description, err)
		return
	}

	for _, additionalTest := range test.additionalTests {
		if err := additionalTest.test(result); err != nil {
			t.Errorf("expected \"%v\" additional test \"%v\" to succeed but it did not: %v", test.description, additionalTest.description, err)
		}
	}
}

func expectFromJSONFailures(t *testing.T, tests []fromJSONTest, testFunc func([]byte) (interface{}, error)) {
	for _, test := range tests {
		expectFromJSONFailure(t, &test, testFunc)
	}
}

func expectFromJSONFailure(t *testing.T, test *fromJSONTest, testFunc func([]byte) (interface{}, error)) {
	if _, err := testFunc(test.bytes); err == nil {
		t.Errorf("expected \"%v\" to fail but it did not", test.description)
	}
}

func quoted(s string) []byte {
	return []byte(fmt.Sprintf("\"%v\"", s))
}
