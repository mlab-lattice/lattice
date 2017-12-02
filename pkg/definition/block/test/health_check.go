package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockHealthCheckHttp() *block.ComponentHealthCheck {
	return &block.ComponentHealthCheck{
		Http: MockHttpHealthCheck(),
	}
}

func MockHealthCheckHttpExpectedJson() []byte {
	return GenerateHealthCheckExpectedJson(
		MockHttpHealthCheckExpectedJson(),
		nil,
		nil,
	)
}

func MockHealthCheckTcp() *block.ComponentHealthCheck {
	return &block.ComponentHealthCheck{
		Tcp: MockTcpHealthCheck(),
	}
}

func MockHealthCheckTcpExpectedJson() []byte {
	return GenerateHealthCheckExpectedJson(
		nil,
		MockTcpHealthCheckExpectedJson(),
		nil,
	)
}

func MockHealthCheckExec() *block.ComponentHealthCheck {
	return &block.ComponentHealthCheck{
		Exec: MockExecHealthCheck(),
	}
}

func MockHealthCheckExecExpectedJson() []byte {
	return GenerateHealthCheckExpectedJson(
		nil,
		nil,
		MockExecHealthCheckExpectedJson(),
	)
}

func GenerateHealthCheckExpectedJson(http, tcp, exec []byte) []byte {
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

func MockHttpHealthCheck() *block.HttpComponentHealthCheck {
	return MockHttpHealthCheckNoHeaders()
}

func MockHttpHealthCheckExpectedJson() []byte {
	return MockHttpHealthCheckNoHeadersExpectedJson()
}

func MockHttpHealthCheckNoHeaders() *block.HttpComponentHealthCheck {
	return &block.HttpComponentHealthCheck{
		Path: "/status",
		Port: "http",
	}
}

func MockHttpHealthCheckNoHeadersExpectedJson() []byte {
	return GenerateHttpHealthCheckExpectedJson(
		[]byte(`"/status"`),
		[]byte(`"http"`),
		nil,
	)
}

func MockHttpHealthCheckWithHeaders() *block.HttpComponentHealthCheck {
	return &block.HttpComponentHealthCheck{
		Path: "/status",
		Port: "http",
		Headers: map[string]string{
			"x-my-header":   "foo",
			"x-your-header": "bar",
		},
	}
}

func MockHttpHealthCheckWithHeadersExpectedJson() []byte {
	return GenerateHttpHealthCheckExpectedJson(
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

func GenerateHttpHealthCheckExpectedJson(path, port, headers []byte) []byte {
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

func MockTcpHealthCheck() *block.TcpComponentHealthCheck {
	return &block.TcpComponentHealthCheck{
		Port: "tcp",
	}
}

func MockTcpHealthCheckExpectedJson() []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "port",
			Bytes: []byte(`"tcp"`),
		},
	})
}

func MockExecHealthCheck() *block.ExecComponentHealthCheck {
	return &block.ExecComponentHealthCheck{
		Command: []string{"./start", "--my-app"},
	}
}

func MockExecHealthCheckExpectedJson() []byte {
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
