package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockComponentBuild() *block.ComponentBuild {
	return MockComponentDockerImageBuild()
}

func MockComponentBuildExpectedJson() []byte {
	return MockDockerImageComponentBuildExpectedJson()
}

func MockComponentDockerImageBuild() *block.ComponentBuild {
	return &block.ComponentBuild{
		DockerImage: MockDockerImage(),
	}
}

func MockDockerImageComponentBuildExpectedJson() []byte {
	return GenerateComponentBuildExpectedJson(
		nil,
		nil,
		nil,
		nil,
		MockDockerImageExpectedJson(),
	)
}

func MockGitRepositoryLanguageComponentBuild() *block.ComponentBuild {
	language := "foobar"
	command := "install"
	return &block.ComponentBuild{
		GitRepository: MockGitRepository(),
		Language:      &language,
		Command:       &command,
	}
}

func MockGitRepositoryLanguageComponentBuildExpectedJson() []byte {
	return GenerateComponentBuildExpectedJson(
		MockGitRepositoryExpectedJson(),
		[]byte(`"foobar"`),
		nil,
		[]byte(`"install"`),
		nil,
	)
}

func MockGitRepositoryBaseDockerImageComponentBuild() *block.ComponentBuild {
	command := "install"
	return &block.ComponentBuild{
		GitRepository:   MockGitRepository(),
		BaseDockerImage: MockDockerImage(),
		Command:         &command,
	}
}

func MockGitRepositoryBaseDockerImageComponentBuildExpectedJson() []byte {
	return GenerateComponentBuildExpectedJson(
		MockGitRepositoryExpectedJson(),
		nil,
		MockDockerImageExpectedJson(),
		[]byte(`"install"`),
		nil,
	)
}

func GenerateComponentBuildExpectedJson(
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
