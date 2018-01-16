package mock

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func ComponentBuild() *block.ComponentBuild {
	return ComponentDockerImageBuild()
}

func ComponentBuildExpectedJSON() []byte {
	return DockerImageComponentBuildExpectedJSON()
}

func ComponentDockerImageBuild() *block.ComponentBuild {
	return &block.ComponentBuild{
		DockerImage: DockerImage(),
	}
}

func DockerImageComponentBuildExpectedJSON() []byte {
	return GenerateComponentBuildExpectedJSON(
		nil,
		nil,
		nil,
		nil,
		DockerImageExpectedJSON(),
	)
}

func GitRepositoryLanguageComponentBuild() *block.ComponentBuild {
	language := "foobar"
	command := "install"
	return &block.ComponentBuild{
		GitRepository: GitRepository(),
		Language:      &language,
		Command:       &command,
	}
}

func GitRepositoryLanguageComponentBuildExpectedJSON() []byte {
	return GenerateComponentBuildExpectedJSON(
		GitRepositoryExpectedJSON(),
		[]byte(`"foobar"`),
		nil,
		[]byte(`"install"`),
		nil,
	)
}

func GitRepositoryBaseDockerImageComponentBuild() *block.ComponentBuild {
	command := "install"
	return &block.ComponentBuild{
		GitRepository:   GitRepository(),
		BaseDockerImage: DockerImage(),
		Command:         &command,
	}
}

func GitRepositoryBaseDockerImageComponentBuildExpectedJSON() []byte {
	return GenerateComponentBuildExpectedJSON(
		GitRepositoryExpectedJSON(),
		nil,
		DockerImageExpectedJSON(),
		[]byte(`"install"`),
		nil,
	)
}

func GenerateComponentBuildExpectedJSON(
	gitRepository,
	language,
	baseDockerImage,
	command,
	dockerImage []byte,
) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:      "git_repository",
			Bytes:     gitRepository,
			OmitEmpty: true,
		},
		{
			Name:      "language",
			Bytes:     language,
			OmitEmpty: true,
		},
		{
			Name:      "base_docker_image",
			Bytes:     baseDockerImage,
			OmitEmpty: true,
		},
		{
			Name:      "command",
			Bytes:     command,
			OmitEmpty: true,
		},
		{
			Name:      "docker_image",
			Bytes:     dockerImage,
			OmitEmpty: true,
		},
	})
}
