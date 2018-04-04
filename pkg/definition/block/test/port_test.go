package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/lattice/pkg/definition/block"
)

func TestPort_Validate(t *testing.T) {
	Validate(
		t,
		nil,

		// Invalid Builds
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.ComponentPort{},
			},
			{
				Description: "no Protocol",
				DefinitionBlock: &block.ComponentPort{
					Name: "foo",
					Port: 1234,
				},
			},
			{
				Description: "no ComponentPort",
				DefinitionBlock: &block.ComponentPort{
					Name:     "foo",
					Protocol: block.ProtocolHTTP,
				},
			},
			{
				Description: "invalid Protocol",
				DefinitionBlock: &block.ComponentPort{
					Name:     "foo",
					Port:     1234,
					Protocol: "invalid",
				},
			},
			{
				Description: "ComponentPort too small",
				DefinitionBlock: &block.ComponentPort{
					Name:     "foo",
					Port:     0,
					Protocol: block.ProtocolHTTP,
				},
			},
			{
				Description: "ComponentPort too large",
				DefinitionBlock: &block.ComponentPort{
					Name:     "foo",
					Port:     100000,
					Protocol: block.ProtocolHTTP,
				},
			},
			{
				Description: "Empty name",
				DefinitionBlock: &block.ComponentPort{
					Port:     1234,
					Protocol: block.ProtocolHTTP,
				},
			},
		},

		// Valid Builds
		[]ValidateTest{
			{
				Description: "Protocol HTTP, ExternalAccess nil",
				DefinitionBlock: &block.ComponentPort{
					Name:     "foo",
					Port:     1234,
					Protocol: block.ProtocolHTTP,
				},
			},
			{
				Description: "ComponentPort, Protocol HTTP, public ExternalAccess",
				DefinitionBlock: &block.ComponentPort{
					Name:           "foo",
					Port:           1234,
					Protocol:       block.ProtocolHTTP,
					ExternalAccess: MockPublicExternalAccess(),
				},
			},
		},
	)
}

func TestPort_JSON(t *testing.T) {
	JSON(
		t,
		reflect.TypeOf(block.ComponentPort{}),
		[]JSONTest{
			{
				Description: "MockPrivateHTTPPort",
				Bytes:       MockPrivateHTTPPortExpectedJSON(),
				ValuePtr:    MockPrivateHTTPPort(),
			},
			{
				Description: "MockPublicHTTPPort",
				Bytes:       MockPublicHTTPPortExpectedJSON(),
				ValuePtr:    MockPublicHTTPPort(),
			},
		},
	)
}

func TestExternalAccess_Validate(t *testing.T) {
	Validate(
		t,
		nil,

		// Invalid Builds
		[]ValidateTest{},

		// Valid Builds
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.ExternalAccess{},
			},
			{
				Description: "Public false",
				DefinitionBlock: &block.ExternalAccess{
					Public: false,
				},
			},
			{
				Description: "Public true",
				DefinitionBlock: &block.ExternalAccess{
					Public: true,
				},
			},
		},
	)
}

func TestExternalAccess_JSON(t *testing.T) {
	JSON(
		t,
		reflect.TypeOf(block.ExternalAccess{}),
		[]JSONTest{
			{
				Description: "MockPublicExternalAccess",
				Bytes:       MockPublicExternalAccessExpectedJSON(),
				ValuePtr:    MockPublicExternalAccess(),
			},
		},
	)
}
