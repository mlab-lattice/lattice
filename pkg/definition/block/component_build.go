package block

import (
	"errors"
	"fmt"
)

type ComponentBuild struct {
	GitRepository   *GitRepository `json:"git_repository,omitempty"`
	Language        *string        `json:"language,omitempty"`
	BaseDockerImage *DockerImage   `json:"base_docker_image,omitempty"`
	Command         *string        `json:"command,omitempty"`
	DockerImage     *DockerImage   `json:"docker_image,omitempty"`
}

// Validate implements Interface
func (cb *ComponentBuild) Validate(interface{}) error {
	if cb.GitRepository != nil && cb.DockerImage != nil {
		return errors.New("cannot specify both git_repository and docker_image")
	}

	if cb.GitRepository != nil {
		return cb.validateGitRepositoryBuild()
	}

	if cb.DockerImage != nil {
		return cb.validateDockerImageBuild()
	}

	return errors.New("must specify either git_repository or docker_image")
}

func (cb *ComponentBuild) validateGitRepositoryBuild() error {
	if err := cb.GitRepository.Validate(nil); err != nil {
		return fmt.Errorf("git_repository definition error: %v", err)
	}

	if cb.Command == nil {
		return errors.New("command is required")
	}

	if cb.Language == nil && cb.BaseDockerImage == nil {
		return errors.New("must specify either language or base_docker_image")
	}

	if cb.Language != nil && cb.BaseDockerImage != nil {
		return errors.New("cannot specify both language and base_docker_image")
	}

	// TODO: potentially validate language here

	if cb.BaseDockerImage != nil {
		if err := cb.BaseDockerImage.Validate(nil); err != nil {
			return fmt.Errorf("base_docker_image definition error: %v", err)
		}
	}

	return nil
}

func (cb *ComponentBuild) validateDockerImageBuild() error {
	if err := cb.DockerImage.Validate(nil); err != nil {
		return fmt.Errorf("docker_image definition error: %v", err)
	}

	if cb.Language != nil {
		return errors.New("cannot specify both docker_image and language")
	}

	if cb.BaseDockerImage != nil {
		return errors.New("cannot specify both docker_image and base_docker_image")
	}

	return nil
}
