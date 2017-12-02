package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

func TestPort_Validate(t *testing.T) {
	TestValidate(
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
					Protocol: block.HttpProtocol,
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
					Protocol: block.HttpProtocol,
				},
			},
			{
				Description: "ComponentPort too large",
				DefinitionBlock: &block.ComponentPort{
					Name:     "foo",
					Port:     100000,
					Protocol: block.HttpProtocol,
				},
			},
			{
				Description: "Empty name",
				DefinitionBlock: &block.ComponentPort{
					Port:     1234,
					Protocol: block.HttpProtocol,
				},
			},
		},

		// Valid Builds
		[]ValidateTest{
			{
				Description: "Protocol Http, ExternalAccess nil",
				DefinitionBlock: &block.ComponentPort{
					Name:     "foo",
					Port:     1234,
					Protocol: block.HttpProtocol,
				},
			},
			{
				Description: "Protocol Tcp, ExternalAccess nil",
				DefinitionBlock: &block.ComponentPort{
					Name:     "foo",
					Port:     1234,
					Protocol: block.TcpProtocol,
				},
			},
			{
				Description: "ComponentPort, Protocol Http, public ExternalAccess",
				DefinitionBlock: &block.ComponentPort{
					Name:           "foo",
					Port:           1234,
					Protocol:       block.HttpProtocol,
					ExternalAccess: MockPublicExternalAccess(),
				},
			},
		},
	)
}

func TestPort_JSON(t *testing.T) {
	TestJSON(
		t,
		reflect.TypeOf(block.ComponentPort{}),
		[]JSONTest{
			{
				Description: "MockPrivateHttpPort",
				Bytes:       MockPrivateHttpPortExpectedJson(),
				ValuePtr:    MockPrivateHttpPort(),
			},
			{
				Description: "MockPublicHttpPort",
				Bytes:       MockPublicHttpPortExpectedJson(),
				ValuePtr:    MockPublicHttpPort(),
			},
			{
				Description: "MockPrivateTcpPort",
				Bytes:       MockPrivateTcpPortExpectedJson(),
				ValuePtr:    MockPrivateTcpPort(),
			},
			{
				Description: "MockPublicTcpPort",
				Bytes:       MockPublicTcpPortExpectedJson(),
				ValuePtr:    MockPublicTcpPort(),
			},
		},
	)
}

func TestExternalAccess_Validate(t *testing.T) {
	TestValidate(
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
	TestJSON(
		t,
		reflect.TypeOf(block.ExternalAccess{}),
		[]JSONTest{
			{
				Description: "MockPublicExternalAccess",
				Bytes:       MockPublicExternalAccessExpectedJson(),
				ValuePtr:    MockPublicExternalAccess(),
			},
		},
	)
}
