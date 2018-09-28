package containerbuilder

import (
	"github.com/mlab-lattice/lattice/pkg/definition/v1"
)

// XXX <GEB>: should probably be merged with b.retrieveSource
func (b *Builder) retrieveLocation(location *v1.Location) (string, error) {
	if location.GitRepository != nil {
		return b.retrieveGitRepository(location.GitRepository)
	}

	return "", newErrorUser("docker build did not include a location")
}
