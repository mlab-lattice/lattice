package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

func TestExec_Validate(t *testing.T) {
	TestValidate(
		t,
		nil,

		// Invalid Exec
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.ComponentExec{},
			},
			{
				Description: "empty Command",
				DefinitionBlock: &block.ComponentExec{
					Command: []string{},
				},
			},
		},

		// Valid Exec
		[]ValidateTest{
			{
				Description: "single Command element",
				DefinitionBlock: &block.ComponentExec{
					Command: []string{"./start"},
				},
			},
			{
				Description: "multiple Command elements",
				DefinitionBlock: &block.ComponentExec{
					Command: []string{"./start", "--my-app", "--now"},
				},
			},
		},
	)
}

func TestExec_JSON(t *testing.T) {
	TestJSON(
		t,
		reflect.TypeOf(block.ComponentExec{}),
		[]JSONTest{
			{
				Description: "MockExec",
				Bytes:       MockExecExpectedJson(),
				ValuePtr:    MockExec(),
			},
		},
	)
}
