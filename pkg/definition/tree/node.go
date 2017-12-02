package tree

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition"
)

// The Node interface represents a Node in the tree of a System definition.
// Note that Nodes are assumed to have an Immutable location in the tree,
// i.e. their parent and children will not change.
type Node interface {
	Parent() Node
	Path() NodePath
	Name() string
	Definition() definition.Interface
	Subsystems() map[NodePath]Node
	Services() map[NodePath]*ServiceNode
}

func NewNode(d definition.Interface, parent Node) (Node, error) {
	var node Node

	// Dispatch to create proper node
	switch st := d.Metadata().Type; st {
	case definition.SystemType:
		sd, ok := d.(*definition.System)
		if !ok {
			return nil, fmt.Errorf("definition.Interface with $.type %v was not a definitions.System", definition.SystemType)
		}

		sn, err := NewSystemNode(sd, parent)
		if err != nil {
			return nil, err
		}

		node = Node(sn)
	case definition.ServiceType:
		sd, ok := d.(*definition.Service)
		if !ok {
			return nil, fmt.Errorf("definition.Interface with $.type %v was not a definitions.System", definition.ServiceType)
		}

		sn, err := NewServiceNode(sd, parent)
		if err != nil {
			return nil, err
		}

		node = Node(sn)
	default:
		// TODO: add TemplateNode
		return nil, fmt.Errorf("invalid $.type %v", st)
	}

	return node, nil
}

func getPath(parent Node, definition definition.Interface) NodePath {
	parentPath := ""
	if parent != nil {
		parentPath = string(parent.Path())
	}

	return NodePath(parentPath + "/" + definition.Metadata().Name)
}
