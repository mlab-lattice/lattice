package componentbuild

import (
	"os"

	systemdefinitionblock "github.com/mlab-lattice/core/pkg/system/definition/block"
	gitutil "github.com/mlab-lattice/core/pkg/util/git"

	dockerclient "github.com/docker/docker/client"
)

type Builder struct {
	WorkingDir          string
	ComponentBuildBlock *systemdefinitionblock.ComponentBuild
	DockerOptions       *DockerOptions
	DockerClient        *dockerclient.Client
	GitResolver         *gitutil.Resolver
	GitResolverOptions  *GitResolverOptions
}

type DockerOptions struct {
	Registry     string
	Repository   string
	Tag          string
	Push         bool
	RegistryAuth *string
}

type GitResolverOptions struct {
	SSHKey []byte
}

type ErrorUser struct {
	message string
}

func newErrorUser(message string) *ErrorUser {
	return &ErrorUser{
		message: message,
	}
}

func (e *ErrorUser) Error() string {
	return e.message
}

type ErrorInternal struct {
	message string
}

func newErrorInternal(message string) *ErrorInternal {
	return &ErrorInternal{
		message: message,
	}
}

func (e *ErrorInternal) Error() string {
	return e.message
}

func NewBuilder(
	workDirectory string,
	dockerOptions *DockerOptions,
	gitResolverOptions *GitResolverOptions,
	componentBuildBlock *systemdefinitionblock.ComponentBuild,
) (*Builder, error) {
	if workDirectory == "" {
		return nil, newErrorInternal("workDirectory not supplied")
	}

	if dockerOptions == nil {
		return nil, newErrorInternal("dockerOptions not supplied")
	}

	if gitResolverOptions == nil {
		gitResolverOptions = &GitResolverOptions{}
	}

	if componentBuildBlock == nil {
		return nil, newErrorInternal("componentBuildBlock not supplied")
	}

	if err := componentBuildBlock.Validate(nil); err != nil {
		return nil, newErrorUser("invalid component build: " + err.Error())
	}

	dockerClient, err := dockerclient.NewEnvClient()
	if err != nil {
		return nil, newErrorInternal("error getting docker client: " + err.Error())
	}

	b := &Builder{
		WorkingDir:          workDirectory,
		ComponentBuildBlock: componentBuildBlock,
		DockerOptions:       dockerOptions,
		DockerClient:        dockerClient,
		GitResolverOptions:  gitResolverOptions,
	}
	return b, nil
}

func (b *Builder) Build() error {
	err := os.MkdirAll(b.WorkingDir, 0777)
	if err != nil {
		return newErrorInternal("failed to create working directory: " + err.Error())
	}

	if b.ComponentBuildBlock.GitRepository != nil {
		return b.buildGitRepositoryComponent()
	}

	return newErrorUser("unsupported component build type")
}
