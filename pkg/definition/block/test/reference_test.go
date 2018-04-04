package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/lattice/pkg/definition/block"
)

func TestReference_JSON(t *testing.T) {
	JSON(
		t,
		reflect.TypeOf(block.Reference{}),
		[]JSONTest{
			{
				Description: "MockReference",
				Bytes:       MockReferenceExpectedJSON(),
				ValuePtr:    MockReference(),
			},
		},
	)
}
