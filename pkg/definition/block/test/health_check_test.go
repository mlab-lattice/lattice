package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

func TestHealthCheck_Validate(t *testing.T) {
	httpHealthCheck := MockHttpHealthCheck()
	tcpHealthCheck := MockTcpHealthCheck()

	TestValidate(
		t,
		map[string]*block.ComponentPort{},

		// Invalid HealthChecks
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.ComponentHealthCheck{},
			},

			// HttpComponentHealthCheck
			{
				Description: "empty HttpComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					Http: &block.HttpComponentHealthCheck{},
				},
			},
			{
				Description: "HttpComponentHealthCheck with invalid ComponentPort",
				DefinitionBlock: &block.ComponentHealthCheck{
					Http: httpHealthCheck,
				},
			},
			{
				Description: "HttpComponentHealthCheck with invalid ComponentPort Protocol",
				DefinitionBlock: &block.ComponentHealthCheck{
					Http: httpHealthCheck,
				},
				Information: map[string]*block.ComponentPort{
					httpHealthCheck.Port: MockTcpPort(),
				},
			},

			// TcpComponentHealthCheck
			{
				Description: "empty TcpComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					Tcp: &block.TcpComponentHealthCheck{},
				},
			},
			{
				Description: "TcpComponentHealthCheck with invalid ComponentPort",
				DefinitionBlock: &block.ComponentHealthCheck{
					Tcp: tcpHealthCheck,
				},
			},
			{
				Description: "TcpComponentHealthCheck with invalid ComponentPort Protocol",
				DefinitionBlock: &block.ComponentHealthCheck{
					Tcp: tcpHealthCheck,
				},
				Information: map[string]*block.ComponentPort{
					tcpHealthCheck.Port: MockHttpPort(),
				},
			},

			// ExecComponentHealthCheck
			{
				Description: "empty ExecComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					Exec: &block.ExecComponentHealthCheck{},
				},
			},

			// Multiple ComponentHealthCheck types
			{
				Description: "HttpComponentHealthCheck, TcpComponentHealthCheck, and ExecComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					Http: httpHealthCheck,
					Tcp:  tcpHealthCheck,
					Exec: MockExecHealthCheck(),
				},
			},
			{
				Description: "HttpComponentHealthCheck and TcpComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					Http: httpHealthCheck,
					Tcp:  tcpHealthCheck,
				},
				Information: map[string]*block.ComponentPort{
					httpHealthCheck.Port: MockHttpPort(),
					tcpHealthCheck.Port:  MockTcpPort(),
				},
			},
			{
				Description: "HttpComponentHealthCheck and ExecComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					Http: httpHealthCheck,
					Exec: MockExecHealthCheck(),
				},
				Information: map[string]*block.ComponentPort{
					httpHealthCheck.Port: MockHttpPort(),
				},
			},
			{
				Description: "TcpComponentHealthCheck and ExecComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					Tcp:  tcpHealthCheck,
					Exec: MockExecHealthCheck(),
				},
				Information: map[string]*block.ComponentPort{
					tcpHealthCheck.Port: MockTcpPort(),
				},
			},
		},

		// Valid HealthChecks
		[]ValidateTest{
			{
				Description: "HttpComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					Http: httpHealthCheck,
				},
				Information: map[string]*block.ComponentPort{
					httpHealthCheck.Port: MockHttpPort(),
				},
			},
			{
				Description: "TcpComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					Tcp: tcpHealthCheck,
				},
				Information: map[string]*block.ComponentPort{
					tcpHealthCheck.Port: MockTcpPort(),
				},
			},
			{
				Description: "ExecComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					Exec: MockExecHealthCheck(),
				},
			},
		},
	)
}

func TestHealthCheck_JSON(t *testing.T) {
	TestJSON(
		t,
		reflect.TypeOf(block.ComponentHealthCheck{}),
		[]JSONTest{
			{
				Description: "MockHealthCheckHttp",
				Bytes:       MockHealthCheckHttpExpectedJson(),
				ValuePtr:    MockHealthCheckHttp(),
			},
			{
				Description: "MockHealthCheckTcp",
				Bytes:       MockHealthCheckTcpExpectedJson(),
				ValuePtr:    MockHealthCheckTcp(),
			},
			{
				Description: "MockHealthCheckExec",
				Bytes:       MockHealthCheckExecExpectedJson(),
				ValuePtr:    MockHealthCheckExec(),
			},
		},
	)
}

func TestHttpHealthCheck_Validate(t *testing.T) {
	TestValidate(
		t,
		map[string]*block.ComponentPort{},

		// Invalid HttpHealthChecks
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.HttpComponentHealthCheck{},
			},
			{
				Description: "invalid ComponentPort and empty Path",
				DefinitionBlock: &block.HttpComponentHealthCheck{
					Port: "http",
				},
				Information: map[string]*block.ComponentPort{
					"http": MockHttpPort(),
				},
			},
			{
				Description: "invalid Path",
				DefinitionBlock: &block.HttpComponentHealthCheck{
					Port: "http",
					Path: "foo",
				},
				Information: map[string]*block.ComponentPort{
					"http": MockHttpPort(),
				},
			},
			{
				Description: "invalid ComponentPort Protocol",
				DefinitionBlock: &block.HttpComponentHealthCheck{
					Port: "http",
					Path: "/status",
				},
				Information: map[string]*block.ComponentPort{
					"http": MockTcpPort(),
				},
			},
		},

		// Valid HttpHealthChecks
		[]ValidateTest{
			{
				Description: "No Headers",
				DefinitionBlock: &block.HttpComponentHealthCheck{
					Port: "http",
					Path: "/status",
				},
				Information: map[string]*block.ComponentPort{
					"http": MockHttpPort(),
				},
			},
			{
				Description: "Headers",
				DefinitionBlock: &block.HttpComponentHealthCheck{
					Port: "http",
					Path: "/status",
					Headers: map[string]string{
						"foo": "bar",
						"biz": "baz",
					},
				},
				Information: map[string]*block.ComponentPort{
					"http": MockHttpPort(),
				},
			},
		},
	)
}

func TestHttpHealthCheck_JSON(t *testing.T) {
	TestJSON(
		t,
		reflect.TypeOf(block.HttpComponentHealthCheck{}),
		[]JSONTest{
			{
				Description: "MockHttpHealthCheckNoHeaders",
				Bytes:       MockHttpHealthCheckNoHeadersExpectedJson(),
				ValuePtr:    MockHttpHealthCheckNoHeaders(),
			},
			{
				Description: "MockHttpHealthCheckWithHeaders",
				Bytes:       MockHttpHealthCheckWithHeadersExpectedJson(),
				ValuePtr:    MockHttpHealthCheckWithHeaders(),
			},
		},
	)
}

func TestTcpHealthCheck_Validate(t *testing.T) {
	TestValidate(
		t,
		map[string]*block.ComponentPort{},

		// Invalid TcpHealthChecks
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.TcpComponentHealthCheck{},
			},
			{
				Description: "invalid ComponentPort",
				DefinitionBlock: &block.TcpComponentHealthCheck{
					Port: "tcp",
				},
			},
			{
				Description: "invalid ComponentPort Protocol",
				DefinitionBlock: &block.TcpComponentHealthCheck{
					Port: "tcp",
				},
				Information: map[string]*block.ComponentPort{
					"tcp": MockHttpPort(),
				},
			},
		},

		// Valid TcpHealthChecks
		[]ValidateTest{
			{
				Description: "Valid ComponentPort",
				DefinitionBlock: &block.TcpComponentHealthCheck{
					Port: "tcp",
				},
				Information: map[string]*block.ComponentPort{
					"tcp": MockTcpPort(),
				},
			},
		},
	)
}

func TestTcpHealthCheck_JSON(t *testing.T) {
	TestJSON(
		t,
		reflect.TypeOf(block.TcpComponentHealthCheck{}),
		[]JSONTest{
			{
				Description: "MockTcpHealthCheck",
				Bytes:       MockTcpHealthCheckExpectedJson(),
				ValuePtr:    MockTcpHealthCheck(),
			},
		},
	)
}

func TestExecHealthCheck_Validate(t *testing.T) {
	TestValidate(
		t,
		nil,

		// Invalid ExecComponentHealthCheck
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.ExecComponentHealthCheck{},
			},
			{
				Description: "empty Command",
				DefinitionBlock: &block.ExecComponentHealthCheck{
					Command: []string{},
				},
			},
		},

		// Valid TcpHealthChecks
		[]ValidateTest{
			{
				Description: "Command with single element",
				DefinitionBlock: &block.ExecComponentHealthCheck{
					Command: []string{"./start"},
				},
			},
			{
				Description: "Command with multiple elements",
				DefinitionBlock: &block.ExecComponentHealthCheck{
					Command: []string{"./start", "--my-app", "--now"},
				},
			},
		},
	)
}

func TestExecHealthCheck_JSON(t *testing.T) {
	TestJSON(
		t,
		reflect.TypeOf(block.ExecComponentHealthCheck{}),
		[]JSONTest{
			{
				Description: "MockExecHealthCheck",
				Bytes:       MockExecHealthCheckExpectedJson(),
				ValuePtr:    MockExecHealthCheck(),
			},
		},
	)
}
