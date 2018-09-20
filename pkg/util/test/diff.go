package test

import (
	"encoding/json"
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func ErrorDiffs(expected, actual string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(expected, actual, true)

	return fmt.Sprintf(
		"Expected result: %v\nActual result: %v\nDiff: %v",
		expected,
		actual,
		fmt.Sprintf(dmp.DiffPrettyText(diffs)),
	)
}

func ErrorDiffsJSON(expected, actual interface{}) string {
	dataExpected, _ := json.Marshal(expected)
	dataActual, _ := json.Marshal(actual)
	return ErrorDiffs(string(dataExpected), string(dataActual))
}
