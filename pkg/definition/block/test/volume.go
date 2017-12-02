package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockVolume() *block.Volume {
	return &block.Volume{
		Name:     "read-write",
		SizeInGb: 512,
	}
}

func MockVolumeExpectedJson() []byte {
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

func MockVolumeMount() *block.ComponentVolumeMount {
	return MockVolumeMountReadOnlyFalse()
}

func MockVolumeMountExpectedJson() []byte {
	return MockVolumeMountReadOnlyFalseExpectedJson()
}

func MockVolumeMountReadOnlyFalse() *block.ComponentVolumeMount {
	return &block.ComponentVolumeMount{
		Name:       "read-write",
		MountPoint: "/foobar",
		ReadOnly:   false,
	}
}

func MockVolumeMountReadOnlyFalseExpectedJson() []byte {
	return GenerateVolumeMountExpectedJson(
		[]byte(`"read-write"`),
		[]byte(`"/foobar"`),
		[]byte(`false`),
	)
}

func MockVolumeMountReadOnlyTrue() *block.ComponentVolumeMount {
	return &block.ComponentVolumeMount{
		Name:       "read-only",
		MountPoint: "/foobar",
		ReadOnly:   true,
	}
}

func MockVolumeMountReadOnlyTrueExpectedJson() []byte {
	return GenerateVolumeMountExpectedJson(
		[]byte(`"read-only"`),
		[]byte(`"/foobar"`),
		[]byte(`true`),
	)
}

func GenerateVolumeMountExpectedJson(name, mountPoint, readOnly []byte) []byte {
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
