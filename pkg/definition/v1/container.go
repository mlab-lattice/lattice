package v1

type Container struct {
	Build *ContainerBuild `json:"build,omitempty"`
	Exec  *ContainerExec  `json:"exec,omitempty"`

	Port  *ContainerPort           `json:"port,omitempty"`
	Ports map[string]ContainerPort `json:"ports,omitempty"`

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

type ContainerEnvironment map[string]*ContainerEnvironmentVariable

type ContainerEnvironmentVariable struct {
	Value  *string
	Secret *string
}

type ContainerPort struct {
	Port           int32          `json:"port"`
	Protocol       string         `json:"protocol"`
	ExternalAccess *ContainerPort `json:"external_access,omitempty"`
}

type ContainerPortExternalAccess struct {
	Public bool `json:"public"`
}

type ContainerHealthCheck struct {
	HTTP *ContainerHealthCheckHTTP `json:"http,omitempty"`
}

type ContainerHealthCheckHTTP struct {
	Path string `json:"path"`
	Port string `json:"port"`
}

type ContainerResources struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}
