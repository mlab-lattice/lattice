package test

import (
	"reflect"
	"testing"

	"encoding/json"
	"fmt"
	"github.com/mlab-lattice/system/pkg/definition/block"
)

func TestSecret_JSON(t *testing.T) {
	fmt.Printf("JSON: %v\n", string(MockSecretReferenceExpectedJSON()))
	fmt.Printf("struct: %#v\n", MockSecretReference())
	sr := MockSecretReference()
	data, err := json.Marshal(&sr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("unmarshalled: %v\n", string(data))
	JSON(
		t,
		reflect.TypeOf(block.SecretValue{}),
		[]JSONTest{
			{
				Description: "MockSecret",
				Bytes:       MockSecretExpectedJSON(),
				ValuePtr:    MockSecret(),
			},
			{
				Description: "MockSecretReference",
				Bytes:       MockSecretReferenceExpectedJSON(),
				ValuePtr:    MockSecretReference(),
			},
		},
	)
}
