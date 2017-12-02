package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockDockerImage() *block.DockerImage {
	return &block.DockerImage{
		Registry:   "registry.company.com",
		Repository: "foobar",
		Tag:        "v1.0.0",
	}
}

func MockDockerImageExpectedJson() []byte {
	return GenerateDockerImageExpectedJson(
		[]byte(`"registry.company.com"`),
		[]byte(`"foobar"`),
		[]byte(`"v1.0.0"`),
	)
}

func GenerateDockerImageExpectedJson(
	registry,
	repository,
	tag []byte,
) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "registry",
			Bytes: registry,
		},
		{
			Name:  "repository",
			Bytes: repository,
		},
		{
			Name:  "tag",
			Bytes: tag,
		},
	})
}
