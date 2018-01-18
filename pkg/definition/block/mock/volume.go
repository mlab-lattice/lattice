package mock

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func Volume() *block.Volume {
	return &block.Volume{
		Name:     "read-write",
		SizeInGb: 512,
	}
}

func VolumeExpectedJSON() []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "name",
			Bytes: []byte(`"read-write"`),
		},
		{
			Name:  "size_in_gb",
			Bytes: []byte(`512`),
		},
	})
}

func VolumeMount() *block.ComponentVolumeMount {
	return VolumeMountReadOnlyFalse()
}

func VolumeMountExpectedJSON() []byte {
	return VolumeMountReadOnlyFalseExpectedJSON()
}

func VolumeMountReadOnlyFalse() *block.ComponentVolumeMount {
	return &block.ComponentVolumeMount{
		Name:       "read-write",
		MountPoint: "/foobar",
		ReadOnly:   false,
	}
}

func VolumeMountReadOnlyFalseExpectedJSON() []byte {
	return GenerateVolumeMountExpectedJSON(
		[]byte(`"read-write"`),
		[]byte(`"/foobar"`),
		[]byte(`false`),
	)
}

func VolumeMountReadOnlyTrue() *block.ComponentVolumeMount {
	return &block.ComponentVolumeMount{
		Name:       "read-only",
		MountPoint: "/foobar",
		ReadOnly:   true,
	}
}

func VolumeMountReadOnlyTrueExpectedJSON() []byte {
	return GenerateVolumeMountExpectedJSON(
		[]byte(`"read-only"`),
		[]byte(`"/foobar"`),
		[]byte(`true`),
	)
}

func GenerateVolumeMountExpectedJSON(name, mountPoint, readOnly []byte) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "name",
			Bytes: name,
		},
		{
			Name:  "mount_point",
			Bytes: mountPoint,
		},
		{
			Name:  "read_only",
			Bytes: readOnly,
		},
	})
}
