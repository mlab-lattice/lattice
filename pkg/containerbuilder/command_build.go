package containerbuilder

import (
	"github.com/mlab-lattice/lattice/pkg/definition/v1"
)

func (b *Builder) buildCommandBuildContainer(commandBuild *v1.ContainerBuildCommand) error {
	sourceDirectory, err := b.retrieveSource(commandBuild.Source)
	if err != nil {
		return err
	}

	baseImage, err := getDockerImageFQNFromDockerImageBlock(&commandBuild.BaseImage)
	if err != nil {
		return err
	}

	// docker needs the build args to be a map from strings to pointer to strings,
	// but commandBuild.Environment maps to strings, so create a buildArgs map
	buildArgs := make(map[string]*string, len(commandBuild.Environment))
	for k, v := range commandBuild.Environment {
		buildArgs[k] = &v
	}
	return b.buildDockerImage(sourceDirectory, baseImage, commandBuild.Command, buildArgs)
}
