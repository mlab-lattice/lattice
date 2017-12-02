package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockExec() *block.ComponentExec {
	return &block.ComponentExec{
		Command: []string{"./start", "--my-app"},
		Environment: map[string]string{
			"biz": "baz",
			"foo": "bar",
		},
	}
}

func MockExecExpectedJson() []byte {
	return GenerateExecExpectedJson(
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

func GenerateExecExpectedJson(
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
