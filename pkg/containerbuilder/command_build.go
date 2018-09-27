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

	return b.buildDockerImage(sourceDirectory, baseImage, commandBuild.Command, commandBuild.Environment)
}
