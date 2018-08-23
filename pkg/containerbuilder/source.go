package containerbuilder

import (
	"github.com/mlab-lattice/lattice/pkg/definition/v1"
)

func (b *Builder) retrieveSource(source *v1.ContainerBuildSource) (string, error) {
	if source.GitRepository != nil {
		return b.retrieveGitRepository(source.GitRepository)
	}

	return "", newErrorUser("container build did not include a source")
}
