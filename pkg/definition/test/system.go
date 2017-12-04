package test

import (
	"github.com/mlab-lattice/system/pkg/definition"
	blocktest "github.com/mlab-lattice/system/pkg/definition/block/test"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockSystem() *definition.System {
	service := MockService()
	def := definition.Interface(service)
	return &definition.System{
		Meta:       *blocktest.MockSystemMetadata(),
		Subsystems: []definition.Interface{def},
	}
}

func MockSystemExpectedJSON() []byte {
	return GenerateSystemExpectedJSON(
		blocktest.MockSystemMetadataExpectedJSON(),
		jsonutil.GenerateArrayBytes([][]byte{
			MockServiceExpectedJSON(),
		}),
	)
}

func GenerateSystemExpectedJSON(metadata, subsystems []byte) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "$",
			Bytes: metadata,
		},
		{
			Name:  "subsystems",
			Bytes: subsystems,
		},
	})
}
