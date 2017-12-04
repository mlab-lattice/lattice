package test

import (
	"github.com/mlab-lattice/system/pkg/definition"
	sdb "github.com/mlab-lattice/system/pkg/definition/block"
	sdbt "github.com/mlab-lattice/system/pkg/definition/block/test"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockService() *definition.Service {
	return &definition.Service{
		Meta:       *sdbt.MockServiceMetadata(),
		Components: []*sdb.Component{sdbt.MockComponent()},
		Resources:  *sdbt.MockResources(),
	}
}

func MockServiceExpectedJSON() []byte {
	return GenerateServiceExpectedJSON(
		sdbt.MockServiceMetadataExpectedJSON(),
		nil,
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.MockComponentExpectedJSON(),
		}),
		sdbt.MockResourcesExpectedJSON(),
	)
}

func MockServiceDifferentName() *definition.Service {
	return &definition.Service{
		Meta:       *sdbt.MockServiceDifferentNameMetadata(),
		Components: []*sdb.Component{sdbt.MockComponent()},
		Resources:  *sdbt.MockResources(),
	}
}

func MockServiceDifferentNameExpectedJSON() []byte {
	return GenerateServiceExpectedJSON(
		sdbt.MockServiceDifferentNameMetadataExpectedJSON(),
		nil,
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.MockComponentExpectedJSON(),
		}),
		sdbt.MockResourcesExpectedJSON(),
	)
}

func MockServiceWithVolume() *definition.Service {
	return &definition.Service{
		Meta:       *sdbt.MockServiceMetadata(),
		Volumes:    []*sdb.Volume{sdbt.MockVolume()},
		Components: []*sdb.Component{sdbt.MockComponentWithVolumeMount()},
		Resources:  *sdbt.MockResources(),
	}
}

func MockServiceWithVolumeExpectedJSON() []byte {
	return GenerateServiceExpectedJSON(
		sdbt.MockServiceMetadataExpectedJSON(),
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.MockVolumeExpectedJSON(),
		}),
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.MockComponentWithVolumeMountExpectedJSON(),
		}),
		sdbt.MockResourcesExpectedJSON(),
	)
}

func GenerateServiceExpectedJSON(
	metadata,
	volumes,
	components,
	resources []byte,
) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "$",
			Bytes: metadata,
		},
		{
			Name:      "volumes",
			Bytes:     volumes,
			OmitEmpty: true,
		},
		{
			Name:  "components",
			Bytes: components,
		},
		{
			Name:  "resources",
			Bytes: resources,
		},
	})
}
