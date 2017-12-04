package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

func TestHealthCheck_Validate(t *testing.T) {
	httpHealthCheck := MockHTTPHealthCheck()
	tcpHealthCheck := MockTCPHealthCheck()

	Validate(
		t,
		map[string]*block.ComponentPort{},

		// Invalid HealthChecks
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.ComponentHealthCheck{},
			},

			// HTTPComponentHealthCheck
			{
				Description: "empty HTTPComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					HTTP: &block.HTTPComponentHealthCheck{},
				},
			},
			{
				Description: "HTTPComponentHealthCheck with invalid ComponentPort",
				DefinitionBlock: &block.ComponentHealthCheck{
					HTTP: httpHealthCheck,
				},
			},
			{
				Description: "HTTPComponentHealthCheck with invalid ComponentPort Protocol",
				DefinitionBlock: &block.ComponentHealthCheck{
					HTTP: httpHealthCheck,
				},
				Information: map[string]*block.ComponentPort{
					httpHealthCheck.Port: MockTCPPort(),
				},
			},

			// TCPComponentHealthCheck
			{
				Description: "empty TCPComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					TCP: &block.TCPComponentHealthCheck{},
				},
			},
			{
				Description: "TCPComponentHealthCheck with invalid ComponentPort",
				DefinitionBlock: &block.ComponentHealthCheck{
					TCP: tcpHealthCheck,
				},
			},
			{
				Description: "TCPComponentHealthCheck with invalid ComponentPort Protocol",
				DefinitionBlock: &block.ComponentHealthCheck{
					TCP: tcpHealthCheck,
				},
				Information: map[string]*block.ComponentPort{
					tcpHealthCheck.Port: MockHTTPPort(),
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
				Description: "HTTPComponentHealthCheck, TCPComponentHealthCheck, and ExecComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					HTTP: httpHealthCheck,
					TCP:  tcpHealthCheck,
					Exec: MockExecHealthCheck(),
				},
			},
			{
				Description: "HTTPComponentHealthCheck and TCPComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					HTTP: httpHealthCheck,
					TCP:  tcpHealthCheck,
				},
				Information: map[string]*block.ComponentPort{
					httpHealthCheck.Port: MockHTTPPort(),
					tcpHealthCheck.Port:  MockTCPPort(),
				},
			},
			{
				Description: "HTTPComponentHealthCheck and ExecComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					HTTP: httpHealthCheck,
					Exec: MockExecHealthCheck(),
				},
				Information: map[string]*block.ComponentPort{
					httpHealthCheck.Port: MockHTTPPort(),
				},
			},
			{
				Description: "TCPComponentHealthCheck and ExecComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					TCP:  tcpHealthCheck,
					Exec: MockExecHealthCheck(),
				},
				Information: map[string]*block.ComponentPort{
					tcpHealthCheck.Port: MockTCPPort(),
				},
			},
		},

		// Valid HealthChecks
		[]ValidateTest{
			{
				Description: "HTTPComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					HTTP: httpHealthCheck,
				},
				Information: map[string]*block.ComponentPort{
					httpHealthCheck.Port: MockHTTPPort(),
				},
			},
			{
				Description: "TCPComponentHealthCheck",
				DefinitionBlock: &block.ComponentHealthCheck{
					TCP: tcpHealthCheck,
				},
				Information: map[string]*block.ComponentPort{
					tcpHealthCheck.Port: MockTCPPort(),
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
	JSON(
		t,
		reflect.TypeOf(block.ComponentHealthCheck{}),
		[]JSONTest{
			{
				Description: "MockHealthCheckHTTP",
				Bytes:       MockHealthCheckHTTPExpectedJSON(),
				ValuePtr:    MockHealthCheckHTTP(),
			},
			{
				Description: "MockHealthCheckTCP",
				Bytes:       MockHealthCheckTCPExpectedJSON(),
				ValuePtr:    MockHealthCheckTCP(),
			},
			{
				Description: "MockHealthCheckExec",
				Bytes:       MockHealthCheckExecExpectedJSON(),
				ValuePtr:    MockHealthCheckExec(),
			},
		},
	)
}

func TestHTTPHealthCheck_Validate(t *testing.T) {
	Validate(
		t,
		map[string]*block.ComponentPort{},

		// Invalid HTTPHealthChecks
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.HTTPComponentHealthCheck{},
			},
			{
				Description: "invalid ComponentPort and empty Path",
				DefinitionBlock: &block.HTTPComponentHealthCheck{
					Port: "http",
				},
				Information: map[string]*block.ComponentPort{
					"http": MockHTTPPort(),
				},
			},
			{
				Description: "invalid Path",
				DefinitionBlock: &block.HTTPComponentHealthCheck{
					Port: "http",
					Path: "foo",
				},
				Information: map[string]*block.ComponentPort{
					"http": MockHTTPPort(),
				},
			},
			{
				Description: "invalid ComponentPort Protocol",
				DefinitionBlock: &block.HTTPComponentHealthCheck{
					Port: "http",
					Path: "/status",
				},
				Information: map[string]*block.ComponentPort{
					"http": MockTCPPort(),
				},
			},
		},

		// Valid HTTPHealthChecks
		[]ValidateTest{
			{
				Description: "No Headers",
				DefinitionBlock: &block.HTTPComponentHealthCheck{
					Port: "http",
					Path: "/status",
				},
				Information: map[string]*block.ComponentPort{
					"http": MockHTTPPort(),
				},
			},
			{
				Description: "Headers",
				DefinitionBlock: &block.HTTPComponentHealthCheck{
					Port: "http",
					Path: "/status",
					Headers: map[string]string{
						"foo": "bar",
						"biz": "baz",
					},
				},
				Information: map[string]*block.ComponentPort{
					"http": MockHTTPPort(),
				},
			},
		},
	)
}

func TestHTTPHealthCheck_JSON(t *testing.T) {
	JSON(
		t,
		reflect.TypeOf(block.HTTPComponentHealthCheck{}),
		[]JSONTest{
			{
				Description: "MockHTTPHealthCheckNoHeaders",
				Bytes:       MockHTTPHealthCheckNoHeadersExpectedJSON(),
				ValuePtr:    MockHTTPHealthCheckNoHeaders(),
			},
			{
				Description: "MockHTTPHealthCheckWithHeaders",
				Bytes:       MockHTTPHealthCheckWithHeadersExpectedJSON(),
				ValuePtr:    MockHTTPHealthCheckWithHeaders(),
			},
		},
	)
}

func TestTCPHealthCheck_Validate(t *testing.T) {
	Validate(
		t,
		map[string]*block.ComponentPort{},

		// Invalid TCPHealthChecks
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.TCPComponentHealthCheck{},
			},
			{
				Description: "invalid ComponentPort",
				DefinitionBlock: &block.TCPComponentHealthCheck{
					Port: "tcp",
				},
			},
			{
				Description: "invalid ComponentPort Protocol",
				DefinitionBlock: &block.TCPComponentHealthCheck{
					Port: "tcp",
				},
				Information: map[string]*block.ComponentPort{
					"tcp": MockHTTPPort(),
				},
			},
		},

		// Valid TCPHealthChecks
		[]ValidateTest{
			{
				Description: "Valid ComponentPort",
				DefinitionBlock: &block.TCPComponentHealthCheck{
					Port: "tcp",
				},
				Information: map[string]*block.ComponentPort{
					"tcp": MockTCPPort(),
				},
			},
		},
	)
}

func TestTCPHealthCheck_JSON(t *testing.T) {
	JSON(
		t,
		reflect.TypeOf(block.TCPComponentHealthCheck{}),
		[]JSONTest{
			{
				Description: "MockTCPHealthCheck",
				Bytes:       MockTCPHealthCheckExpectedJSON(),
				ValuePtr:    MockTCPHealthCheck(),
			},
		},
	)
}

func TestExecHealthCheck_Validate(t *testing.T) {
	Validate(
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

		// Valid TCPHealthChecks
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
	JSON(
		t,
		reflect.TypeOf(block.ExecComponentHealthCheck{}),
		[]JSONTest{
			{
				Description: "MockExecHealthCheck",
				Bytes:       MockExecHealthCheckExpectedJSON(),
				ValuePtr:    MockExecHealthCheck(),
			},
		},
	)
}
