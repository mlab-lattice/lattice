package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/lattice/pkg/definition/block"
)

func TestComponent_Validate(t *testing.T) {
	Validate(
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
	JSON(
		t,
		reflect.TypeOf(block.Component{}),
		[]JSONTest{
			{
				Description: "MockComponent",
				Bytes:       MockComponentExpectedJSON(),
				ValuePtr:    MockComponent(),
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
