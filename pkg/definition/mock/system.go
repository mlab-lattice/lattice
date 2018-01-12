package mock

import (
	"github.com/mlab-lattice/system/pkg/definition"
	blocktest "github.com/mlab-lattice/system/pkg/definition/block/test"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func SystemExpectedJSON() []byte {
	return GenerateSystemExpectedJSON(
		blocktest.SystemMetadataExpectedJSON(),
		jsonutil.GenerateArrayBytes([][]byte{
			ServiceExpectedJSON(),
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
