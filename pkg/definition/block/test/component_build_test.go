package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

func TestBuild_Validate(t *testing.T) {
	language := "foobar"
	command := "install"
	Validate(
		t,
		nil,

		// Invalid Builds
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.ComponentBuild{},
			},
			{
				Description: "both GitRepository and DockerImage",
				DefinitionBlock: &block.ComponentBuild{
					GitRepository: &block.GitRepository{},
					DockerImage:   &block.DockerImage{},
				},
			},
			{
				Description: "GitRepository and no BaseDockerImage or Language",
				DefinitionBlock: &block.ComponentBuild{
					GitRepository: MockGitRepository(),
				},
			},
			{
				Description: "GitRepository and both BaseDockerImage and Language",
				DefinitionBlock: &block.ComponentBuild{
					GitRepository:   MockGitRepository(),
					Language:        &language,
					BaseDockerImage: MockDockerImage(),
				},
			},
			{
				Description: "GitRepository and Language, no Command",
				DefinitionBlock: &block.ComponentBuild{
					GitRepository: MockGitRepository(),
					Language:      &language,
				},
			},
			{
				Description: "GitRepository and BaseDockerImage, no Command",
				DefinitionBlock: &block.ComponentBuild{
					GitRepository:   MockGitRepository(),
					BaseDockerImage: MockDockerImage(),
				},
			},
			{
				Description: "GitRepository and Command, no BaseDockerImage or Language",
				DefinitionBlock: &block.ComponentBuild{
					GitRepository: MockGitRepository(),
					Command:       &command,
				},
			},
			{
				Description: "DockerImage and Language",
				DefinitionBlock: &block.ComponentBuild{
					DockerImage: MockDockerImage(),
					Language:    &language,
				},
			},
			{
				Description: "DockerImage and BaseDockerImage",
				DefinitionBlock: &block.ComponentBuild{
					DockerImage:     MockDockerImage(),
					BaseDockerImage: MockDockerImage(),
				},
			},
			{
				Description: "DockerImage, Language, and BaseDockerImage",
				DefinitionBlock: &block.ComponentBuild{
					DockerImage:     MockDockerImage(),
					Language:        &language,
					BaseDockerImage: MockDockerImage(),
				},
			},
		},

		// Valid Builds
		[]ValidateTest{
			{
				Description: "GitRepository and Language",
				DefinitionBlock: &block.ComponentBuild{
					GitRepository: MockGitRepository(),
					Language:      &language,
					Command:       &command,
				},
			},
			{
				Description: "GitRepository and BaseDockerImage",
				DefinitionBlock: &block.ComponentBuild{
					GitRepository:   MockGitRepository(),
					BaseDockerImage: MockDockerImage(),
					Command:         &command,
				},
			},
			{
				Description: "DockerImage",
				DefinitionBlock: &block.ComponentBuild{
					DockerImage: MockDockerImage(),
				},
			},
		},
	)
}

func TestBuild_JSON(t *testing.T) {
	JSON(
		t,
		reflect.TypeOf(block.ComponentBuild{}),
		[]JSONTest{
			{
				Description: "MockComponentDockerImageBuild",
				Bytes:       MockDockerImageComponentBuildExpectedJSON(),
				ValuePtr:    MockComponentDockerImageBuild(),
			},
			{
				Description: "MockGitRepositoryLanguageComponentBuild",
				Bytes:       MockGitRepositoryLanguageComponentBuildExpectedJSON(),
				ValuePtr:    MockGitRepositoryLanguageComponentBuild(),
			},
			{
				Description: "MockGitRepositoryBaseDockerImageComponentBuild",
				Bytes:       MockGitRepositoryBaseDockerImageComponentBuildExpectedJSON(),
				ValuePtr:    MockGitRepositoryBaseDockerImageComponentBuild(),
			},
		},
	)
}
