package v1

const (
	DockerBuildDefaultPath = "."
)

type DockerImage struct {
	Registry   string `json:"registry,omitempty"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

type DockerBuildContext struct {
	Location *Location `json:"location,omitempty"`
	Path     string    `json:"path,omitempty"`
}

type DockerFile struct {
	Location *Location `json:"location,omitempty"`
	Path     string    `json:"path,omitempty"`
}

type DockerBuildOptions map[string]ValueOrSecret
type DockerBuildEnvironment map[string]ValueOrSecret

type DockerBuild struct {
	BuildContext *DockerBuildContext    `json:"build_context,omitempty"`
	DockerFile   *DockerFile            `json:"docker_file,omitempty"`
	Environment  DockerBuildEnvironment `json:"environment,omitempty"`
	Options      DockerBuildOptions     `json:"options,omitempty"`
}
