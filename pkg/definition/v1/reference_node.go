package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type ReferenceNode struct {
	parent    tree.ComponentNode
	path      tree.NodePath
	reference *Reference
}

func NewReferenceNode(reference *Reference, name string, parent tree.ComponentNode) *ReferenceNode {
	return &ReferenceNode{
		parent:    parent,
		path:      parent.Path().Child(name),
		reference: reference,
	}
}

func (n *ReferenceNode) Parent() tree.ComponentNode {
	return n.parent
}

func (n *ReferenceNode) Path() tree.NodePath {
	return n.path
}

func (n *ReferenceNode) Component() component.Interface {
	return n.reference
}

func (n *ReferenceNode) Reference() *Reference {
	return n.reference
}
