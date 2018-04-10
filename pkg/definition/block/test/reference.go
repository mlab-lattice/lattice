package test

import (
	"github.com/mlab-lattice/lattice/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/lattice/pkg/util/json"
)

func MockReference() *block.Reference {
	return &block.Reference{
		Path: "a.b.c",
	}
}

func MockReferenceExpectedJSON() []byte {
	return GenerateReferenceExpectedJSON([]byte("\"a.b.c\""))
}

func GenerateReferenceExpectedJSON(path []byte) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "reference",
			Bytes: path,
		},
	})
}
