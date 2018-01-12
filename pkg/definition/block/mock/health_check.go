package mock

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func HealthCheckHTTP() *block.ComponentHealthCheck {
	return &block.ComponentHealthCheck{
		HTTP: HTTPHealthCheck(),
	}
}

func HealthCheckHTTPExpectedJSON() []byte {
	return GenerateHealthCheckExpectedJSON(
		HTTPHealthCheckExpectedJSON(),
		nil,
		nil,
	)
}

func HealthCheckTCP() *block.ComponentHealthCheck {
	return &block.ComponentHealthCheck{
		TCP: TCPHealthCheck(),
	}
}

func HealthCheckTCPExpectedJSON() []byte {
	return GenerateHealthCheckExpectedJSON(
		nil,
		TCPHealthCheckExpectedJSON(),
		nil,
	)
}

func HealthCheckExec() *block.ComponentHealthCheck {
	return &block.ComponentHealthCheck{
		Exec: ExecHealthCheck(),
	}
}

func HealthCheckExecExpectedJSON() []byte {
	return GenerateHealthCheckExpectedJSON(
		nil,
		nil,
		ExecHealthCheckExpectedJSON(),
	)
}

func GenerateHealthCheckExpectedJSON(http, tcp, exec []byte) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:      "http",
			Bytes:     http,
			OmitEmpty: true,
		},
		{
			Name:      "tcp",
			Bytes:     tcp,
			OmitEmpty: true,
		},
		{
			Name:      "exec",
			Bytes:     exec,
			OmitEmpty: true,
		},
	})
}

func HTTPHealthCheck() *block.HTTPComponentHealthCheck {
	return HTTPHealthCheckNoHeaders()
}

func HTTPHealthCheckExpectedJSON() []byte {
	return HTTPHealthCheckNoHeadersExpectedJSON()
}

func HTTPHealthCheckNoHeaders() *block.HTTPComponentHealthCheck {
	return &block.HTTPComponentHealthCheck{
		Path: "/status",
		Port: "http",
	}
}

func HTTPHealthCheckNoHeadersExpectedJSON() []byte {
	return GenerateHTTPHealthCheckExpectedJSON(
		[]byte(`"/status"`),
		[]byte(`"http"`),
		nil,
	)
}

func HTTPHealthCheckWithHeaders() *block.HTTPComponentHealthCheck {
	return &block.HTTPComponentHealthCheck{
		Path: "/status",
		Port: "http",
		Headers: map[string]string{
			"x-my-header":   "foo",
			"x-your-header": "bar",
		},
	}
}

func HTTPHealthCheckWithHeadersExpectedJSON() []byte {
	return GenerateHTTPHealthCheckExpectedJSON(
		[]byte(`"/status"`),
		[]byte(`"http"`),
		jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
			{
				Name:  "x-my-header",
				Bytes: []byte(`"foo"`),
			},
			{
				Name:  "x-your-header",
				Bytes: []byte(`"bar"`),
			},
		}),
	)
}

func GenerateHTTPHealthCheckExpectedJSON(path, port, headers []byte) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "path",
			Bytes: path,
		},
		{
			Name:  "port",
			Bytes: port,
		},
		{
			Name:      "headers",
			Bytes:     headers,
			OmitEmpty: true,
		},
	})
}

func TCPHealthCheck() *block.TCPComponentHealthCheck {
	return &block.TCPComponentHealthCheck{
		Port: "tcp",
	}
}

func TCPHealthCheckExpectedJSON() []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "port",
			Bytes: []byte(`"tcp"`),
		},
	})
}

func ExecHealthCheck() *block.ExecComponentHealthCheck {
	return &block.ExecComponentHealthCheck{
		Command: []string{"./start", "--my-app"},
	}
}

func ExecHealthCheckExpectedJSON() []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name: "command",
			Bytes: jsonutil.GenerateArrayBytes([][]byte{
				[]byte(`"./start"`),
				[]byte(`"--my-app"`),
			}),
		},
	})
}
