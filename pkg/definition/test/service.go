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

func MockServiceExpectedJson() []byte {
	return GenerateServiceExpectedJson(
		sdbt.MockServiceMetadataExpectedJson(),
		nil,
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.MockComponentExpectedJson(),
		}),
		sdbt.MockResourcesExpectedJson(),
	)
}

func MockServiceDifferentName() *definition.Service {
	return &definition.Service{
		Meta:       *sdbt.MockServiceDifferentNameMetadata(),
		Components: []*sdb.Component{sdbt.MockComponent()},
		Resources:  *sdbt.MockResources(),
	}
}

func MockServiceDifferentNameExpectedJson() []byte {
	return GenerateServiceExpectedJson(
		sdbt.MockServiceDifferentNameMetadataExpectedJson(),
		nil,
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.MockComponentExpectedJson(),
		}),
		sdbt.MockResourcesExpectedJson(),
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

func MockServiceWithVolumeExpectedJson() []byte {
	return GenerateServiceExpectedJson(
		sdbt.MockServiceMetadataExpectedJson(),
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.MockVolumeExpectedJson(),
		}),
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.MockComponentWithVolumeMountExpectedJson(),
		}),
		sdbt.MockResourcesExpectedJson(),
	)
}

func GenerateServiceExpectedJson(
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
