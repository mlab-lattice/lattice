package test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func ErrorDiffs(t *testing.T, description, expected, actual string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(expected, actual, true)

	t.Errorf(
		"%v:\nExpected result: %v\nActual result: %v\nDiff: %v",
		description,
		expected,
		actual,
		fmt.Sprintf(dmp.DiffPrettyText(diffs)),
	)
}

func ValidateToJson(t *testing.T, description string, actualMarshaler interface{}, expected []byte) {
	actual, err := json.Marshal(actualMarshaler)

	if err != nil {
		t.Fatalf("Error returned when marshalling: %v", err)
	}

	// Have to jump through all these hoops because ordering of maps isn't guaranteed
	var actualUnmarshaled, expectedMarshaled map[string]interface{}
	if err = json.Unmarshal(actual, &actualUnmarshaled); err != nil {
		t.Fatalf("Error returned when unmarshaling actual: %v", err)
	}
	if err = json.Unmarshal(expected, &expectedMarshaled); err != nil {
		t.Fatalf("Error returned when unmarshaling expected: %v", err)
	}

	if !reflect.DeepEqual(actualUnmarshaled, expectedMarshaled) {
		actualRemarshaled, err := json.Marshal(actualUnmarshaled)
		if err != nil {
			t.Fatalf("Error returned rewhen marshalling actual: %v", err)
		}

		expectedRemarshaled, err := json.Marshal(expectedMarshaled)
		if err != nil {
			t.Fatalf("Error returned rewhen marshalling expected: %v", err)
		}
		ErrorDiffs(t, description, string(expectedRemarshaled), string(actualRemarshaled))
	}
}
