package test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition/block"
	testutil "github.com/mlab-lattice/system/pkg/util/test"
)

type ValidateTest struct {
	Description     string
	DefinitionBlock interface{}
	Information     interface{}
}

func TestValidate(
	t *testing.T,
	defaultInformation interface{},
	expectFailureTests,
	expectSuccessTests []ValidateTest,
) {
	for _, test := range expectFailureTests {
		information := test.Information
		if information == nil {
			information = defaultInformation
		}

		ExpectUnsuccessfulValidationWithInformation(
			t,
			test.DefinitionBlock.(block.Interface),
			test.Description,
			information,
		)
	}

	for _, test := range expectSuccessTests {
		information := test.Information
		if information == nil {
			information = defaultInformation
		}

		ExpectSuccessfulValidationWithInformation(
			t,
			test.DefinitionBlock.(block.Interface),
			test.Description,
			information,
		)
	}
}

func ExpectSuccessfulValidation(t *testing.T, d block.Interface, description string) {
	ExpectSuccessfulValidationWithInformation(t, d, description, struct{}{})
}

func ExpectSuccessfulValidationWithInformation(
	t *testing.T,
	d block.Interface,
	description string,
	information interface{},
) {
	err := d.Validate(information)
	if err != nil {
		t.Errorf("Expected no error for %v but got %v", description, err)
	}
}

func ExpectUnsuccessfulValidation(t *testing.T, block block.Interface, description string) {
	ExpectUnsuccessfulValidationWithInformation(t, block, description, struct{}{})
}

func ExpectUnsuccessfulValidationWithInformation(
	t *testing.T,
	d block.Interface,
	description string,
	information interface{},
) {
	err := d.Validate(information)
	if err == nil {
		t.Errorf("No error returned when validating %v", description)
	}
}

type JSONTest struct {
	Description string
	Bytes       []byte
	ValuePtr    interface{}
}

func TestJSON(t *testing.T, valueType reflect.Type, tests []JSONTest) {
	marshalTests := []MarshalJSONTest{}
	unmarshalTests := []UnmarshalJSONTest{}

	for _, test := range tests {
		marshalTests = append(marshalTests, MarshalJSONTest{
			Description:   test.Description,
			BytesProducer: test.ValuePtr,
			ExpectedBytes: test.Bytes,
		})
		unmarshalTests = append(unmarshalTests, UnmarshalJSONTest{
			Description: test.Description,
			Bytes:       test.Bytes,
			ExpectedPtr: test.ValuePtr,
		})
	}

	TestMarshalJSON(t, marshalTests)
	TestUnmarshalJSON(t, valueType, unmarshalTests)
}

type MarshalJSONTest struct {
	Description   string
	BytesProducer interface{}
	ExpectedBytes []byte
}

func TestMarshalJSON(t *testing.T, tests []MarshalJSONTest) {
	for _, test := range tests {
		testutil.ValidateToJson(t, test.Description, test.BytesProducer, test.ExpectedBytes)
	}
}

type UnmarshalJSONTest struct {
	Description string
	Bytes       []byte
	ExpectedPtr interface{}
}

func TestUnmarshalJSON(t *testing.T, expectedType reflect.Type, tests []UnmarshalJSONTest) {
	for _, test := range tests {
		actual := reflect.New(expectedType).Interface()
		if err := json.Unmarshal(test.Bytes, &actual); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(test.ExpectedPtr, actual) {
			testutil.ErrorDiffs(
				t,
				test.Description,
				fmt.Sprintf("%#v", test.ExpectedPtr),
				fmt.Sprintf("%#v", actual),
			)
		}
	}
}
