package test

import (
	"encoding/json"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition"
	sdbt "github.com/mlab-lattice/system/pkg/definition/block/test"
	testutil "github.com/mlab-lattice/system/pkg/util/test"
	"reflect"
)

func TestSystem_Validate(t *testing.T) {
	service := MockService()
	serviceDefinition := definition.Interface(service)

	secondService := MockService()
	secondService.Meta.Name = "my-second-service"
	secondServiceDefinition := definition.Interface(secondService)

	subsystemSystem := MockSystem()
	subsystemSystemDefinition := definition.Interface(subsystemSystem)

	sdbt.TestValidate(
		t,
		nil,

		// Invalid Builds
		[]sdbt.ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &definition.System{},
			},
			{
				Description: "no Subsystems",
				DefinitionBlock: &definition.System{
					Meta: *sdbt.MockSystemMetadata(),
				},
			},
			{
				Description: "Multiple Subsystems with the same Name",
				DefinitionBlock: &definition.System{
					Meta:       *sdbt.MockSystemMetadata(),
					Subsystems: []definition.Interface{serviceDefinition, serviceDefinition},
				},
			},
		},

		// Valid Builds
		[]sdbt.ValidateTest{
			{
				Description: "Single Service Subsystem",
				DefinitionBlock: &definition.System{
					Meta:       *sdbt.MockSystemMetadata(),
					Subsystems: []definition.Interface{serviceDefinition},
				},
			},
			{
				Description: "Single System Subsystem",
				DefinitionBlock: &definition.System{
					Meta:       *sdbt.MockSystemMetadata(),
					Subsystems: []definition.Interface{subsystemSystemDefinition},
				},
			},
			{
				Description: "Multiple Subsystems with different Names",
				DefinitionBlock: &definition.System{
					Meta:       *sdbt.MockSystemMetadata(),
					Subsystems: []definition.Interface{serviceDefinition, secondServiceDefinition},
				},
			},
		},
	)
}

func TestSystem_MarshalJSON(t *testing.T) {
	tests := map[string]struct {
		system        *definition.System
		expectedBytes []byte
	}{
		"MockSystem": {
			system:        MockSystem(),
			expectedBytes: MockSystemExpectedJSON(),
		},
	}

	for description, test := range tests {
		testutil.ValidateToJSON(t, description, test.system, test.expectedBytes)
	}
}

func TestSystem_UnmarshalJSON(t *testing.T) {
	sys := &definition.System{}
	err := json.Unmarshal(MockSystemExpectedJSON(), sys)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(sys, MockSystem()) {
		actualMarshaled, err := json.Marshal(sys)
		if err != nil {
			t.Fatalf("Error returned when marshalling actual: %v", err)
		}

		expectedMarshaled, err := json.Marshal(MockSystem())
		if err != nil {
			t.Fatalf("Error returned when marshalling expected: %v", err)
		}
		testutil.ErrorDiffs(t, "MockSystem", string(actualMarshaled), string(expectedMarshaled))
	}
}
