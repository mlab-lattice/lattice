package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

func TestVolume_Validate(t *testing.T) {
	Validate(
		t,
		nil,

		// Invalid Builds
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.Volume{},
			},
			{
				Description: "empty Name",
				DefinitionBlock: &block.Volume{
					SizeInGb: 10,
				},
			},
			{
				Description: "Volume SizeInGb too large",
				DefinitionBlock: &block.Volume{
					Name:     "foo",
					SizeInGb: 2048,
				},
			},
		},

		// Valid Builds
		[]ValidateTest{
			{
				Description: "Valid",
				DefinitionBlock: &block.Volume{
					Name:     "foo",
					SizeInGb: 10,
				},
			},
		},
	)
}

func TestVolume_JSON(t *testing.T) {
	JSON(
		t,
		reflect.TypeOf(block.Volume{}),
		[]JSONTest{
			{
				Description: "MockVolume",
				Bytes:       MockVolumeExpectedJSON(),
				ValuePtr:    MockVolume(),
			},
		},
	)
}

func TestVolumeMount_Validate(t *testing.T) {
	Validate(
		t,
		nil,

		// Invalid Builds
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.ComponentVolumeMount{},
			},
			{
				Description: "empty Name",
				DefinitionBlock: &block.ComponentVolumeMount{
					MountPoint: "/foo",
				},
			},
			{
				Description: "invalid MountPath",
				DefinitionBlock: &block.ComponentVolumeMount{
					Name:       "foo",
					MountPoint: "foo",
				},
			},
		},

		// Valid Builds
		[]ValidateTest{
			{
				Description: "ReadOnly false",
				DefinitionBlock: &block.ComponentVolumeMount{
					Name:       "foo",
					MountPoint: "/foo",
				},
			},
			{
				Description: "ReadOnly true",
				DefinitionBlock: &block.ComponentVolumeMount{
					Name:       "foo",
					MountPoint: "/foo",
					ReadOnly:   true,
				},
			},
		},
	)
}

func TestVolumeMount_JSON(t *testing.T) {
	JSON(
		t,
		reflect.TypeOf(block.ComponentVolumeMount{}),
		[]JSONTest{
			{
				Description: "MockVolumeMountReadOnlyFalse",
				Bytes:       MockVolumeMountReadOnlyFalseExpectedJSON(),
				ValuePtr:    MockVolumeMountReadOnlyFalse(),
			},
			{
				Description: "MockVolumeMountReadOnlyTrue",
				Bytes:       MockVolumeMountReadOnlyTrueExpectedJSON(),
				ValuePtr:    MockVolumeMountReadOnlyTrue(),
			},
		},
	)
}
