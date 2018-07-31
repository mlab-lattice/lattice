package tree

import (
	"github.com/mlab-lattice/lattice/pkg/definition/component"
)

// The Node interface represents a node in a tree.
type Node interface {
	Path() NodePath
	Value() interface{}
}

// The ComponentNode interface represents a node in the tree of a System definition.
type ComponentNode interface {
	Node
	Component() component.Interface
}
