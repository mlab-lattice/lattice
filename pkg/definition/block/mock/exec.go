package mock

import (
	"github.com/mlab-lattice/lattice/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/lattice/pkg/util/json"
)

func Exec() *block.ComponentExec {
	baz := "baz"
	bar := "baz"
	return &block.ComponentExec{
		Command: []string{"./start", "--my-app"},
		Environment: block.Environment{
			"biz": &block.EnvironmentVariable{
				Value: &baz,
			},
			"foo": &block.EnvironmentVariable{
				Value: &bar,
			},
		},
	}
}

func ExecExpectedJSON() []byte {
	return GenerateExecExpectedJSON(
		jsonutil.GenerateArrayBytes([][]byte{
			[]byte(`"./start"`),
			[]byte(`"--my-app"`),
		}),
		jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
			{
				Name:  "biz",
				Bytes: []byte(`"baz"`),
			},
			{
				Name:  "foo",
				Bytes: []byte(`"bar"`),
			},
		}),
	)
}

func GenerateExecExpectedJSON(
	command,
	environment []byte,
) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "command",
			Bytes: command,
		},
		{
			Name:      "environment",
			Bytes:     environment,
			OmitEmpty: true,
		},
	})
}
