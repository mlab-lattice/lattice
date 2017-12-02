package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockComponent() *block.Component {
	return &block.Component{
		Name:  "service",
		Build: *MockComponentBuild(),
		Exec:  *MockExec(),
	}
}

func MockComponentExpectedJson() []byte {
	return GenerateComponentExpectedJson(
		[]byte(`"service"`),
		[]byte(`false`),
		nil,
		nil,
		MockComponentBuildExpectedJson(),
		MockExecExpectedJson(),
		nil,
	)
}

func MockComponentInitTrue() *block.Component {
	return &block.Component{
		Init:  true,
		Name:  "service",
		Build: *MockComponentBuild(),
		Exec:  *MockExec(),
	}
}

func MockComponentInitTrueExpectedJson() []byte {
	return GenerateComponentExpectedJson(
		[]byte(`"service"`),
		[]byte(`true`),
		nil,
		nil,
		MockComponentBuildExpectedJson(),
		MockExecExpectedJson(),
		nil,
	)
}

func MockComponentWithHttpPort() *block.Component {
	return &block.Component{
		Name:  "service",
		Ports: []*block.ComponentPort{MockHttpPort()},
		Build: *MockComponentBuild(),
		Exec:  *MockExec(),
	}
}

func MockComponentWithHttpPortExpectedJson() []byte {
	return GenerateComponentExpectedJson(
		[]byte(`"service"`),
		[]byte(`false`),
		jsonutil.GenerateArrayBytes([][]byte{MockHttpPortExpectedJson()}),
		nil,
		MockComponentBuildExpectedJson(),
		MockExecExpectedJson(),
		nil,
	)
}

func MockComponentWithHttpPortHttpHealthCheck() *block.Component {
	return &block.Component{
		Name:        "service",
		Ports:       []*block.ComponentPort{MockHttpPort()},
		Build:       *MockComponentBuild(),
		Exec:        *MockExec(),
		HealthCheck: MockHealthCheckHttp(),
	}
}

func MockComponentWithHttpPortHttpHealthCheckExpectedJson() []byte {
	return GenerateComponentExpectedJson(
		[]byte(`"service"`),
		[]byte(`false`),
		jsonutil.GenerateArrayBytes([][]byte{MockHttpPortExpectedJson()}),
		nil,
		MockComponentBuildExpectedJson(),
		MockExecExpectedJson(),
		MockHealthCheckHttpExpectedJson(),
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

func MockComponentWithVolumeMountExpectedJson() []byte {
	return GenerateComponentExpectedJson(
		[]byte(`"service"`),
		[]byte(`false`),
		nil,
		jsonutil.GenerateArrayBytes([][]byte{MockVolumeMountReadOnlyFalseExpectedJson()}),
		MockComponentBuildExpectedJson(),
		MockExecExpectedJson(),
		nil,
	)
}

func GenerateComponentExpectedJson(
	name,
	init,
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
			Name:  "init",
			Bytes: init,
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
