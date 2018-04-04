package test

import (
	"github.com/mlab-lattice/lattice/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/lattice/pkg/util/json"
)

func MockHealthCheckHTTP() *block.ComponentHealthCheck {
	return &block.ComponentHealthCheck{
		HTTP: MockHTTPHealthCheck(),
	}
}

func MockHealthCheckHTTPExpectedJSON() []byte {
	return GenerateHealthCheckExpectedJSON(
		MockHTTPHealthCheckExpectedJSON(),
		nil,
		nil,
	)
}

func MockHealthCheckExec() *block.ComponentHealthCheck {
	return &block.ComponentHealthCheck{
		Exec: MockExecHealthCheck(),
	}
}

func MockHealthCheckExecExpectedJSON() []byte {
	return GenerateHealthCheckExpectedJSON(
		nil,
		nil,
		MockExecHealthCheckExpectedJSON(),
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

func MockHTTPHealthCheck() *block.HTTPComponentHealthCheck {
	return MockHTTPHealthCheckNoHeaders()
}

func MockHTTPHealthCheckExpectedJSON() []byte {
	return MockHTTPHealthCheckNoHeadersExpectedJSON()
}

func MockHTTPHealthCheckNoHeaders() *block.HTTPComponentHealthCheck {
	return &block.HTTPComponentHealthCheck{
		Path: "/status",
		Port: "http",
	}
}

func MockHTTPHealthCheckNoHeadersExpectedJSON() []byte {
	return GenerateHTTPHealthCheckExpectedJSON(
		[]byte(`"/status"`),
		[]byte(`"http"`),
		nil,
	)
}

func MockHTTPHealthCheckWithHeaders() *block.HTTPComponentHealthCheck {
	return &block.HTTPComponentHealthCheck{
		Path: "/status",
		Port: "http",
		Headers: map[string]string{
			"x-my-header":   "foo",
			"x-your-header": "bar",
		},
	}
}

func MockHTTPHealthCheckWithHeadersExpectedJSON() []byte {
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

func MockExecHealthCheck() *block.ExecComponentHealthCheck {
	return &block.ExecComponentHealthCheck{
		Command: []string{"./start", "--my-app"},
	}
}

func MockExecHealthCheckExpectedJSON() []byte {
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
