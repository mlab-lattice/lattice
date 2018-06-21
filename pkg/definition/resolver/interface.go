package resolver

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

// SystemResolver resolves system definitions from different sources such as git
type SystemResolver interface {
	ResolveDefinition(uri string, gitResolveOptions *git.Options) (tree.Node, error)
	ListDefinitionVersions(uri string, gitResolveOptions *git.Options) ([]string, error)
}
