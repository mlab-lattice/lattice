package test

import (
	"github.com/mlab-lattice/lattice/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/lattice/pkg/util/json"
)

func MockComponent() *block.Component {
	return &block.Component{
		Name:  "service",
		Build: *MockComponentBuild(),
		Exec:  *MockExec(),
	}
}

func MockComponentExpectedJSON() []byte {
	return GenerateComponentExpectedJSON(
		[]byte(`"service"`),
		nil,
		nil,
		MockComponentBuildExpectedJSON(),
		MockExecExpectedJSON(),
		nil,
	)
}

func MockComponentWithHTTPPort() *block.Component {
	return &block.Component{
		Name:  "service",
		Ports: []*block.ComponentPort{MockHTTPPort()},
		Build: *MockComponentBuild(),
		Exec:  *MockExec(),
	}
}

func MockComponentWithHTTPPortExpectedJSON() []byte {
	return GenerateComponentExpectedJSON(
		[]byte(`"service"`),
		jsonutil.GenerateArrayBytes([][]byte{MockHTTPPortExpectedJSON()}),
		nil,
		MockComponentBuildExpectedJSON(),
		MockExecExpectedJSON(),
		nil,
	)
}

func MockComponentWithHTTPPortHTTPHealthCheck() *block.Component {
	return &block.Component{
		Name:        "service",
		Ports:       []*block.ComponentPort{MockHTTPPort()},
		Build:       *MockComponentBuild(),
		Exec:        *MockExec(),
		HealthCheck: MockHealthCheckHTTP(),
	}
}

func MockComponentWithHTTPPortHTTPHealthCheckExpectedJSON() []byte {
	return GenerateComponentExpectedJSON(
		[]byte(`"service"`),
		jsonutil.GenerateArrayBytes([][]byte{MockHTTPPortExpectedJSON()}),
		nil,
		MockComponentBuildExpectedJSON(),
		MockExecExpectedJSON(),
		MockHealthCheckHTTPExpectedJSON(),
	)
}

func MockComponentWithVolumeMount() *block.Component {
	return &block.Component{
		Name:         "service",
		VolumeMounts: []*block.ComponentVolumeMount{MockVolumeMountReadOnlyFalse()},
		Build:        *MockComponentBuild(),
		Exec:         *MockExec(),
	}
}

func MockComponentWithVolumeMountExpectedJSON() []byte {
	return GenerateComponentExpectedJSON(
		[]byte(`"service"`),
		nil,
		jsonutil.GenerateArrayBytes([][]byte{MockVolumeMountReadOnlyFalseExpectedJSON()}),
		MockComponentBuildExpectedJSON(),
		MockExecExpectedJSON(),
		nil,
	)
}

func GenerateComponentExpectedJSON(
	name,
	ports,
	volumeMounts,
	build,
	exec,
	healthCheck []byte,
) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "name",
			Bytes: name,
		},
		{
			Name:      "ports",
			Bytes:     ports,
			OmitEmpty: true,
		},
		{
			Name:      "volume_mounts",
			Bytes:     volumeMounts,
			OmitEmpty: true,
		},
		{
			Name:  "build",
			Bytes: build,
		},
		{
			Name:  "exec",
			Bytes: exec,
		},
		{
			Name:      "health_check",
			Bytes:     healthCheck,
			OmitEmpty: true,
		},
	})
}
