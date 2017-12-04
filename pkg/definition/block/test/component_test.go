package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

func TestComponent_Validate(t *testing.T) {
	TestValidate(
		t,
		map[string]*block.Volume{},

		// Invalid Components
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.Component{},
			},
			{
				Description: "ComponentBuild, empty Exec, Name",
				DefinitionBlock: &block.Component{
					Name: "foo",
					Exec: *MockExec(),
				},
			},
			{
				Description: "Exec, empty ComponentBuild, Name",
				DefinitionBlock: &block.Component{
					Exec: *MockExec(),
				},
			},
			{
				Description: "ComponentBuild and Exec, empty Name",
				DefinitionBlock: &block.Component{
					Build: *MockComponentBuild(),
					Exec:  *MockExec(),
				},
			},

			// HTTPComponentHealthCheck
			{
				Description: "HTTPComponentHealthCheck with no Ports",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckHTTP(),
				},
			},
			{
				Description: "HTTPComponentHealthCheck with invalid Ports",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckHTTP(),
					Ports: []*block.ComponentPort{
						{
							Name:     "foo",
							Protocol: block.ProtocolHTTP,
						},
					},
				},
			},
			{
				Description: "HTTPComponentHealthCheck with wrong Protocol",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckHTTP(),
					Ports: []*block.ComponentPort{
						{
							Name:     "http",
							Protocol: block.ProtocolTCP,
						},
					},
				},
			},

			// TCPComponentHealthCheck
			{
				Description: "TCPComponentHealthCheck with no Ports",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckTCP(),
				},
			},
			{
				Description: "TCPComponentHealthCheck with invalid ComponentPort",
				DefinitionBlock: &block.Component{Name: "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckTCP(),
					Ports: []*block.ComponentPort{
						{
							Name:     "foo",
							Protocol: block.ProtocolTCP,
						},
					},
				},
			},
			{
				Description: "TCPComponentHealthCheck with wrong Protocol",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckTCP(),
					Ports: []*block.ComponentPort{
						{
							Name:     "tcp",
							Protocol: block.ProtocolHTTP,
						},
					},
				},
			},

			// ComponentVolumeMount
			{
				Description: "invalid ComponentVolumeMount",
				DefinitionBlock: &block.Component{
					Name:  "foo",
					Build: *MockComponentBuild(),
					Exec:  *MockExec(),
					VolumeMounts: []*block.ComponentVolumeMount{
						{
							Name:       "foo",
							MountPoint: "bar",
						},
					},
				},
			},
		},

		// Valid Builds
		[]ValidateTest{
			{
				Description: "ComponentBuild, Exec and Name",
				DefinitionBlock: &block.Component{
					Name:  "foo",
					Build: *MockComponentBuild(),
					Exec:  *MockExec(),
				},
			},
			{
				Description: "Init",
				DefinitionBlock: &block.Component{
					Name:  "foo",
					Init:  true,
					Build: *MockComponentBuild(),
					Exec:  *MockExec(),
				},
			},
			{
				Description: "ExecComponentHealthCheck",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckExec(),
				},
			},
			{
				Description: "HTTPComponentHealthCheck",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckHTTP(),
					Ports:       []*block.ComponentPort{MockHTTPPort()},
				},
			},
			{
				Description: "TCPComponentHealthCheck",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckTCP(),
					Ports:       []*block.ComponentPort{MockTCPPort()},
				},
			},
			{
				Description: "ComponentVolumeMount",
				DefinitionBlock: &block.Component{
					Name:         "foo",
					Build:        *MockComponentBuild(),
					Exec:         *MockExec(),
					VolumeMounts: []*block.ComponentVolumeMount{MockVolumeMount()},
				},
				Information: map[string]*block.Volume{
					MockVolumeMount().Name: {
						Name:     MockVolumeMount().Name,
						SizeInGb: 50,
					},
				},
			},
		},
	)
}

func TestComponent_JSON(t *testing.T) {
	TestJSON(
		t,
		reflect.TypeOf(block.Component{}),
		[]JSONTest{
			{
				Description: "MockComponent",
				Bytes:       MockComponentExpectedJSON(),
				ValuePtr:    MockComponent(),
			},
			{
				Description: "MockComponentInitTrue",
				Bytes:       MockComponentInitTrueExpectedJSON(),
				ValuePtr:    MockComponentInitTrue(),
			},
			{
				Description: "MockComponentWithHTTPPort",
				Bytes:       MockComponentWithHTTPPortExpectedJSON(),
				ValuePtr:    MockComponentWithHTTPPort(),
			},
			{
				Description: "MockComponentWithHTTPPortHTTPHealthCheck",
				Bytes:       MockComponentWithHTTPPortHTTPHealthCheckExpectedJSON(),
				ValuePtr:    MockComponentWithHTTPPortHTTPHealthCheck(),
			},
			{
				Description: "MockComponentWithVolumeMount",
				Bytes:       MockComponentWithVolumeMountExpectedJSON(),
				ValuePtr:    MockComponentWithVolumeMount(),
			},
		},
	)
}
