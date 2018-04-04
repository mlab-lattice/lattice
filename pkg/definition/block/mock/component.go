package mock

import (
	"github.com/mlab-lattice/lattice/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/lattice/pkg/util/json"
)

func Component() *block.Component {
	return &block.Component{
		Name:  "service",
		Build: *ComponentBuild(),
		Exec:  *Exec(),
	}
}

func ComponentExpectedJSON() []byte {
	return GenerateComponentExpectedJSON(
		[]byte(`"service"`),
		nil,
		nil,
		ComponentBuildExpectedJSON(),
		ExecExpectedJSON(),
		nil,
	)
}

func ComponentWithHTTPPort() *block.Component {
	return &block.Component{
		Name:  "service",
		Ports: []*block.ComponentPort{HTTPPort()},
		Build: *ComponentBuild(),
		Exec:  *Exec(),
	}
}

func ComponentWithHTTPPortExpectedJSON() []byte {
	return GenerateComponentExpectedJSON(
		[]byte(`"service"`),
		jsonutil.GenerateArrayBytes([][]byte{HTTPPortExpectedJSON()}),
		nil,
		ComponentBuildExpectedJSON(),
		ExecExpectedJSON(),
		nil,
	)
}

func ComponentWithHTTPPortHTTPHealthCheck() *block.Component {
	return &block.Component{
		Name:        "service",
		Ports:       []*block.ComponentPort{HTTPPort()},
		Build:       *ComponentBuild(),
		Exec:        *Exec(),
		HealthCheck: HealthCheckHTTP(),
	}
}

func ComponentWithHTTPPortHTTPHealthCheckExpectedJSON() []byte {
	return GenerateComponentExpectedJSON(
		[]byte(`"service"`),
		jsonutil.GenerateArrayBytes([][]byte{HTTPPortExpectedJSON()}),
		nil,
		ComponentBuildExpectedJSON(),
		ExecExpectedJSON(),
		HealthCheckHTTPExpectedJSON(),
	)
}

func ComponentWithVolumeMount() *block.Component {
	return &block.Component{
		Name:         "service",
		VolumeMounts: []*block.ComponentVolumeMount{VolumeMountReadOnlyFalse()},
		Build:        *ComponentBuild(),
		Exec:         *Exec(),
	}
}

func ComponentWithVolumeMountExpectedJSON() []byte {
	return GenerateComponentExpectedJSON(
		[]byte(`"service"`),
		nil,
		jsonutil.GenerateArrayBytes([][]byte{VolumeMountReadOnlyFalseExpectedJSON()}),
		ComponentBuildExpectedJSON(),
		ExecExpectedJSON(),
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
