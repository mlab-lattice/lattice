package mock

import (
	"github.com/mlab-lattice/lattice/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/lattice/pkg/util/json"
)

func DockerImage() *block.DockerImage {
	return &block.DockerImage{
		Registry:   "registry.company.com",
		Repository: "foobar",
		Tag:        "v1.0.0",
	}
}

func DockerImageExpectedJSON() []byte {
	return GenerateDockerImageExpectedJSON(
		[]byte(`"registry.company.com"`),
		[]byte(`"foobar"`),
		[]byte(`"v1.0.0"`),
	)
}

func GenerateDockerImageExpectedJSON(
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
