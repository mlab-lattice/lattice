package v1

import (
	"encoding/json"
	"fmt"
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

type ContainerEnvironment map[string]ContainerEnvironmentVariable

type ContainerEnvironmentVariable struct {
	Value  *string
	Secret *string
}

func (cev ContainerEnvironmentVariable) MarshalJSON() ([]byte, error) {
	if cev.Value != nil {
		e := containerEnvironmentVariableEncoder(*cev.Value)
		return json.Marshal(&e)
	}

	if cev.Secret != nil {
		e := containerEnvironmentVariableSecretEncoder{
			Secret: *cev.Secret,
		}
		return json.Marshal(&e)
	}

	return nil, fmt.Errorf("ContainerEnvironmentVariable must have either value or secret")
}

func (cev *ContainerEnvironmentVariable) UnmarshalJSON(data []byte) error {
	var val containerEnvironmentVariableEncoder
	err := json.Unmarshal(data, &val)
	if err == nil {
		strVal := string(val)
		cev.Value = &strVal
		return nil
	}

	// If the error wasn't that the data wasn't a string, return the error.
	if _, ok := err.(*json.UnmarshalTypeError); !ok {
		return err
	}

	// Otherwise, try to see if the environment variable is a secret
	var secret containerEnvironmentVariableSecretEncoder
	err = json.Unmarshal(data, &secret)
	if err == nil {
		cev.Secret = &secret.Secret
		return nil
	}

	return err
}

type containerEnvironmentVariableEncoder string

type containerEnvironmentVariableSecretEncoder struct {
	Secret string `json:"secret"`
}

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
