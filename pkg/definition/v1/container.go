package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/component"
)

const ComponentTypeContainer = "container"

var ContainerType = component.Type{
	APIVersion: APIVersion,
	Type:       ComponentTypeContainer,
}

type Container struct {
	Build *ContainerBuild `json:"build,omitempty"`
	Exec  *ContainerExec  `json:"exec,omitempty"`

	Ports map[int32]ContainerPort `json:"ports,omitempty"`

	HealthCheck *ContainerHealthCheck `json:"health_check,omitempty"`

	Resources *ContainerResources `json:"resources,omitempty"`
}

type ContainerBuild struct {
	GitRepository   *GitRepository       `json:"git_repository,omitempty"`
	Language        *string              `json:"language,omitempty"`
	BaseDockerImage *DockerImage         `json:"base_docker_image,omitempty"`
	Command         []string             `json:"command,omitempty"`
	Environment     ContainerEnvironment `json:"environment,omitempty"`

	DockerImage *DockerImage `json:"docker_image,omitempty"`
}

type ContainerExec struct {
	Command     []string             `json:"command"`
	Environment ContainerEnvironment `json:"environment,omitempty"`
}

type ContainerEnvironment map[string]ValueOrSecret

type ContainerPort struct {
	Protocol       string                       `json:"protocol"`
	ExternalAccess *ContainerPortExternalAccess `json:"external_access,omitempty"`
}

func (c ContainerPort) Public() bool {
	return c.ExternalAccess != nil && c.ExternalAccess.Public
}

type ContainerPortExternalAccess struct {
	Public bool `json:"public"`
}

type ContainerHealthCheck struct {
	HTTP *ContainerHealthCheckHTTP `json:"http,omitempty"`
}

type ContainerHealthCheckHTTP struct {
	Path string `json:"path"`
	Port int32  `json:"port"`
}

type ContainerResources struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}
