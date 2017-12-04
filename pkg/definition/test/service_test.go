package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/block"
	blocktest "github.com/mlab-lattice/system/pkg/definition/block/test"
)

func TestService_Validate(t *testing.T) {
	blocktest.TestValidate(
		t,
		nil,

		// Invalid Builds
		[]blocktest.ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &definition.Service{},
			},
			{
				Description: "no Components or Resources",
				DefinitionBlock: &definition.Service{
					Meta: *blocktest.MockServiceMetadata(),
				},
			},
			{
				Description: "no Resources",
				DefinitionBlock: &definition.Service{
					Meta:       *blocktest.MockServiceMetadata(),
					Components: []*block.Component{blocktest.MockComponent()},
				},
			},
			{
				Description: "empty Meta",
				DefinitionBlock: &definition.Service{
					Components: []*block.Component{blocktest.MockComponent()},
					Resources:  *blocktest.MockResources(),
				},
			},
		},

		// Valid Builds
		[]blocktest.ValidateTest{
			{
				Description: "stateless",
				DefinitionBlock: &definition.Service{
					Meta:       *blocktest.MockServiceMetadata(),
					Components: []*block.Component{blocktest.MockComponent()},
					Resources:  *blocktest.MockResources(),
				},
			},
			{
				Description: "stateful",
				DefinitionBlock: &definition.Service{
					Meta:       *blocktest.MockServiceMetadata(),
					Components: []*block.Component{blocktest.MockComponent()},
					Resources:  *blocktest.MockResources(),
				},
			},
			{
				Description: "Volume no VolumeMount",
				DefinitionBlock: &definition.Service{
					Meta:       *blocktest.MockServiceMetadata(),
					Components: []*block.Component{blocktest.MockComponent()},
					Resources:  *blocktest.MockResources(),
					Volumes:    []*block.Volume{blocktest.MockVolume()},
				},
			},
			{
				Description: "Volume with VolumeMount",
				DefinitionBlock: &definition.Service{
					Meta:       *blocktest.MockServiceMetadata(),
					Components: []*block.Component{blocktest.MockComponentWithVolumeMount()},
					Resources:  *blocktest.MockResources(),
					Volumes:    []*block.Volume{blocktest.MockVolume()},
				},
			},
		},
	)
}

func TestService_JSON(t *testing.T) {
	blocktest.TestJSON(
		t,
		reflect.TypeOf(definition.Service{}),
		[]blocktest.JSONTest{
			{
				Description: "MockService",
				Bytes:       MockServiceExpectedJSON(),
				ValuePtr:    MockService(),
			},
			{
				Description: "MockServiceDifferentName",
				Bytes:       MockServiceDifferentNameExpectedJSON(),
				ValuePtr:    MockServiceDifferentName(),
			},
			{
				Description: "MockServiceWithVolume",
				Bytes:       MockServiceWithVolumeExpectedJSON(),
				ValuePtr:    MockServiceWithVolume(),
			},
		},
	)
}
