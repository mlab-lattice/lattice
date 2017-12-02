package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

func TestMetadata_Validate(t *testing.T) {
	TestValidate(
		t,
		nil,

		// Invalid Metadata
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.Metadata{},
			},
			{
				Description: "no Type",
				DefinitionBlock: &block.Metadata{
					Name: "my-system",
				},
			},
			{
				Description: "no Name",
				DefinitionBlock: &block.Metadata{
					Type: "my-type",
				},
			},
		},

		// Valid Metadata
		[]ValidateTest{
			{
				Description: "Name and Type",
				DefinitionBlock: &block.Metadata{
					Name: "my-system",
					Type: "my-type",
				},
			},
			{
				Description: "Name, Type, and Description",
				DefinitionBlock: &block.Metadata{
					Name:        "my-system",
					Type:        "my-type",
					Description: "this is my system",
				},
			},
		},
	)
}

func TestMetadata_JSON(t *testing.T) {
	TestJSON(
		t,
		reflect.TypeOf(block.Metadata{}),
		[]JSONTest{
			{
				Description: "MockSystemMetadata",
				Bytes:       MockSystemMetadataExpectedJson(),
				ValuePtr:    MockSystemMetadata(),
			},
			{
				Description: "MockServiceMetadata",
				Bytes:       MockServiceMetadataExpectedJson(),
				ValuePtr:    MockServiceMetadata(),
			},
			{
				Description: "MockServiceDifferentNameMetadata",
				Bytes:       MockServiceDifferentNameMetadataExpectedJson(),
				ValuePtr:    MockServiceDifferentNameMetadata(),
			},
		},
	)
}

// TODO: add MetadataParameter
