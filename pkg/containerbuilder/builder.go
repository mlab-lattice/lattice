package containerbuilder

import (
	"os"

	dockerclient "github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/docker"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

type Builder struct {
	BuildID       v1.ContainerBuildID
	SystemID      v1.SystemID
	WorkingDir    string
	DockerOptions *DockerOptions
	DockerClient  *dockerclient.Client
	GitOptions    *git.Options
	StatusUpdater StatusUpdater
}

type DockerOptions struct {
	Registry             string
	Repository           string
	Tag                  string
	Push                 bool
	RegistryAuthProvider docker.RegistryLoginProvider
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

type Failure struct {
	Error error
	Phase v1.ContainerBuildPhase
}

func NewBuilder(
	buildID v1.ContainerBuildID,
	systemID v1.SystemID,
	workDirectory string,
	dockerOptions *DockerOptions,
	gitResolverOptions *git.Options,
	updater StatusUpdater,
) (*Builder, error) {
	if workDirectory == "" {
		return nil, newErrorInternal("workDirectory not supplied")
	}

	if dockerOptions == nil {
		return nil, newErrorInternal("dockerOptions not supplied")
	}

	if gitResolverOptions == nil {
		gitResolverOptions = &git.Options{}
	}

	dockerClient, err := dockerclient.NewEnvClient()
	if err != nil {
		return nil, newErrorInternal("error getting docker client: " + err.Error())
	}

	// Otherwise color detects it's not actually in a terminal and disables itself
	color.NoColor = false

	b := &Builder{
		BuildID:       buildID,
		SystemID:      systemID,
		WorkingDir:    workDirectory,
		DockerOptions: dockerOptions,
		DockerClient:  dockerClient,
		GitOptions:    gitResolverOptions,
		StatusUpdater: updater,
	}
	return b, nil
}

func (b *Builder) Build(containerBuild *definitionv1.ContainerBuild) error {
	err := os.MkdirAll(b.WorkingDir, 0777)
	if err != nil {
		return newErrorInternal("failed to create working directory: " + err.Error())
	}

	if containerBuild.CommandBuild != nil {
		return b.handleError(b.buildCommandBuildContainer(containerBuild.CommandBuild))
	}

	if containerBuild.DockerImage != nil {
		return b.handleError(b.buildDockerImageContainer(containerBuild.DockerImage))
	}

	return newErrorUser("unsupported container build type")
}

func (b *Builder) handleError(err error) error {
	if err == nil {
		return nil
	}

	color.Red("✘ Failed")

	if b.StatusUpdater == nil {
		return err
	}

	switch err.(type) {
	case *ErrorUser:
		b.StatusUpdater.UpdateError(b.BuildID, b.SystemID, false, err)
	default:
		// TODO: is there a reason to differentiate between an ErrorInternal and a non ErrorUser?
		b.StatusUpdater.UpdateError(b.BuildID, b.SystemID, true, err)
	}

	return err
}
