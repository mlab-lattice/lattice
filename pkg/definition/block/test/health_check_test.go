package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/lattice/pkg/definition/block"
)

func TestHealthCheck_Validate(t *testing.T) {
	httpHealthCheck := MockHTTPHealthCheck()

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
					Exec: MockExecHealthCheck(),
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
