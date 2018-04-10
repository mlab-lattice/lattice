package test

import (
	"github.com/mlab-lattice/lattice/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/lattice/pkg/util/json"
)

func MockVolume() *block.Volume {
	return &block.Volume{
		Name:     "read-write",
		SizeInGb: 512,
	}
}

func MockVolumeExpectedJSON() []byte {
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

func MockVolumeMountExpectedJSON() []byte {
	return MockVolumeMountReadOnlyFalseExpectedJSON()
}

func MockVolumeMountReadOnlyFalse() *block.ComponentVolumeMount {
	return &block.ComponentVolumeMount{
		Name:       "read-write",
		MountPoint: "/foobar",
		ReadOnly:   false,
	}
}

func MockVolumeMountReadOnlyFalseExpectedJSON() []byte {
	return GenerateVolumeMountExpectedJSON(
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

func MockVolumeMountReadOnlyTrueExpectedJSON() []byte {
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
