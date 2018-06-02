package tree

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition"
)

// The Node interface represents a Node in the tree of a System definition.
// Note that Nodes are assumed to have an Immutable location in the tree,
// i.e. their parent and children will not change.
type Node interface {
	Parent() Node
	Path() NodePath
	Subsystems() map[NodePath]Node
	Definition() interface{}
}

func NewNode(def interface{}, parent Node) (Node, error) {
	// Dispatch to create proper node
	switch d := def.(type) {
	case *definition.System:
		return NewSystemNode(d, parent)

	case *definition.Service:
		return NewServiceNode(d, parent)

	default:
		return nil, fmt.Errorf("unrecognized definition struct")
	}
}

func Walk(n Node, fn func(Node) error) error {
	err := fn(n)
	if err != nil {
		return fmt.Errorf("error walking node %v: %v", n.Path().String(), err)
	}

	for _, subsystem := range n.Subsystems() {
		err := Walk(subsystem, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

func getPath(parent Node, name string) NodePath {
	parentPath := ""
	if parent != nil {
		parentPath = string(parent.Path())
	}

	return NodePath(fmt.Sprintf("%v/%v", parentPath, name))
}
