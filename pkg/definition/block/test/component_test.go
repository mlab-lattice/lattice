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

			// HttpComponentHealthCheck
			{
				Description: "HttpComponentHealthCheck with no Ports",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckHttp(),
				},
			},
			{
				Description: "HttpComponentHealthCheck with invalid Ports",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckHttp(),
					Ports: []*block.ComponentPort{
						{
							Name:     "foo",
							Protocol: block.HttpProtocol,
						},
					},
				},
			},
			{
				Description: "HttpComponentHealthCheck with wrong Protocol",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckHttp(),
					Ports: []*block.ComponentPort{
						{
							Name:     "http",
							Protocol: block.TcpProtocol,
						},
					},
				},
			},

			// TcpComponentHealthCheck
			{
				Description: "TcpComponentHealthCheck with no Ports",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckTcp(),
				},
			},
			{
				Description: "TcpComponentHealthCheck with invalid ComponentPort",
				DefinitionBlock: &block.Component{Name: "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckTcp(),
					Ports: []*block.ComponentPort{
						{
							Name:     "foo",
							Protocol: block.TcpProtocol,
						},
					},
				},
			},
			{
				Description: "TcpComponentHealthCheck with wrong Protocol",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckTcp(),
					Ports: []*block.ComponentPort{
						{
							Name:     "tcp",
							Protocol: block.HttpProtocol,
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
				Description: "HttpComponentHealthCheck",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckHttp(),
					Ports:       []*block.ComponentPort{MockHttpPort()},
				},
			},
			{
				Description: "TcpComponentHealthCheck",
				DefinitionBlock: &block.Component{
					Name:        "foo",
					Build:       *MockComponentBuild(),
					Exec:        *MockExec(),
					HealthCheck: MockHealthCheckTcp(),
					Ports:       []*block.ComponentPort{MockTcpPort()},
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
				Bytes:       MockComponentExpectedJson(),
				ValuePtr:    MockComponent(),
			},
			{
				Description: "MockComponentInitTrue",
				Bytes:       MockComponentInitTrueExpectedJson(),
				ValuePtr:    MockComponentInitTrue(),
			},
			{
				Description: "MockComponentWithHttpPort",
				Bytes:       MockComponentWithHttpPortExpectedJson(),
				ValuePtr:    MockComponentWithHttpPort(),
			},
			{
				Description: "MockComponentWithHttpPortHttpHealthCheck",
				Bytes:       MockComponentWithHttpPortHttpHealthCheckExpectedJson(),
				ValuePtr:    MockComponentWithHttpPortHttpHealthCheck(),
			},
			{
				Description: "MockComponentWithVolumeMount",
				Bytes:       MockComponentWithVolumeMountExpectedJson(),
				ValuePtr:    MockComponentWithVolumeMount(),
			},
		},
	)
}
