package tree

import (
	"github.com/mlab-lattice/lattice/pkg/definition/resource"
)

// The Node interface represents a Node in the tree of a System definition.
// Note that Nodes are assumed to have an Immutable location in the tree,
// i.e. their parent and children will not change.
type Node interface {
	Parent() Node
	Path() NodePath
	Resource() resource.Interface
}
