package containerbuilder

import (
	"fmt"
	"strings"

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

	dockerfileCommand := fmt.Sprintf("RUN %v", strings.Join(commandBuild.Command, " "))
	return b.buildDockerImage(sourceDirectory, baseImage, dockerfileCommand)
}
