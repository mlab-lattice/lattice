package resolver

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func NewNode(i *ResolutionInfo, path tree.Path, parent tree.Node, children map[string]tree.Node) *Node {
	return &Node{
		Info:     i,
		path:     path,
		parent:   parent,
		children: children,
	}
}

type Node struct {
	Info     *ResolutionInfo
	path     tree.Path
	parent   tree.Node
	children map[string]tree.Node
}

func (n *Node) Path() tree.Path {
	return n.path
}

func (n *Node) Value() interface{} {
	return n.Info
}

func (n *Node) Parent() tree.Node {
	return n.parent
}

func (n *Node) Children() map[string]tree.Node {
	return n.children
}

func (n *Node) Lookup(p tree.Path) (*Node, bool, error) {
	r, ok, err := tree.Lookup(n, p)
	if err != nil {
		return nil, false, err
	}

	if !ok {
		return nil, false, nil
	}

	rn, ok := r.(*Node)
	if !ok {
		return nil, false, fmt.Errorf("resolver node child %v was not a resolver node", p.String())
	}

	return rn, true, nil
}
