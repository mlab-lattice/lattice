package mock

import (
	"github.com/mlab-lattice/system/pkg/definition"
	sdb "github.com/mlab-lattice/system/pkg/definition/block"
	sdbt "github.com/mlab-lattice/system/pkg/definition/block/test"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func Service() *definition.Service {
	return &definition.Service{
		Meta:       *sdbt.ServiceMetadata(),
		Components: []*sdb.Component{sdbt.Component()},
		Resources:  *sdbt.Resources(),
	}
}

func ServiceExpectedJSON() []byte {
	return GenerateServiceExpectedJSON(
		sdbt.ServiceMetadataExpectedJSON(),
		nil,
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.ComponentExpectedJSON(),
		}),
		sdbt.ResourcesExpectedJSON(),
	)
}

func ServiceDifferentName() *definition.Service {
	return &definition.Service{
		Meta:       *sdbt.ServiceDifferentNameMetadata(),
		Components: []*sdb.Component{sdbt.Component()},
		Resources:  *sdbt.Resources(),
	}
}

func ServiceDifferentNameExpectedJSON() []byte {
	return GenerateServiceExpectedJSON(
		sdbt.ServiceDifferentNameMetadataExpectedJSON(),
		nil,
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.ComponentExpectedJSON(),
		}),
		sdbt.ResourcesExpectedJSON(),
	)
}

func ServiceWithVolume() *definition.Service {
	return &definition.Service{
		Meta:       *sdbt.ServiceMetadata(),
		Volumes:    []*sdb.Volume{sdbt.Volume()},
		Components: []*sdb.Component{sdbt.ComponentWithVolumeMount()},
		Resources:  *sdbt.Resources(),
	}
}

func ServiceWithVolumeExpectedJSON() []byte {
	return GenerateServiceExpectedJSON(
		sdbt.ServiceMetadataExpectedJSON(),
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.VolumeExpectedJSON(),
		}),
		jsonutil.GenerateArrayBytes([][]byte{
			sdbt.ComponentWithVolumeMountExpectedJSON(),
		}),
		sdbt.ResourcesExpectedJSON(),
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
