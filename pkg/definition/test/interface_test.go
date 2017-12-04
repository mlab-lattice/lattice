package test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition"
	testutil "github.com/mlab-lattice/system/pkg/util/test"
)

func TestInterface_NewServiceFromJSON(t *testing.T) {
	tests := []struct {
		Description string
		Expected    *definition.Service
		JSON        []byte
	}{
		{
			Description: "MockService",
			Expected:    MockService(),
			JSON:        MockServiceExpectedJSON(),
		},
		{
			Description: "MockServiceDifferentName",
			Expected:    MockServiceDifferentName(),
			JSON:        MockServiceDifferentNameExpectedJSON(),
		},
		{
			Description: "MockServiceWithVolume",
			Expected:    MockServiceWithVolume(),
			JSON:        MockServiceWithVolumeExpectedJSON(),
		},
	}

	for _, test := range tests {
		def, err := definition.UnmarshalJSON(test.JSON)
		if err != nil {
			t.Fatal(err)
		}

		service, ok := def.(*definition.Service)
		if !ok {
			t.Fatalf("%v: was not a sd.Service", test.Description)
		}

		if !reflect.DeepEqual(service, test.Expected) {
			testutil.ErrorDiffs(
				t,
				test.Description,
				fmt.Sprintf("%#v", test.Expected),
				fmt.Sprintf("%#v", service),
			)
		}
	}
}

func TestInterface_NewSystemFromJSON(t *testing.T) {
	tests := []struct {
		Description string
		Expected    *definition.System
		JSON        []byte
	}{
		{
			Description: "MockSystem",
			Expected:    MockSystem(),
			JSON:        MockSystemExpectedJSON(),
		},
	}

	for _, test := range tests {
		def, err := definition.UnmarshalJSON(test.JSON)
		if err != nil {
			t.Fatal(err)
		}

		system, ok := def.(*definition.System)
		if !ok {
			t.Fatalf("%v: was not a sd.System", test.Description)
		}

		if !reflect.DeepEqual(system, test.Expected) {
			testutil.ErrorDiffs(
				t,
				test.Description,
				fmt.Sprintf("%#v", test.Expected),
				fmt.Sprintf("%#v", system),
			)
		}
	}
}
