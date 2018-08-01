package resolver

import (
	"fmt"

	"encoding/json"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func NewNode(i *ResolutionInfo, path tree.Path, children map[string]tree.Node) *Node {
	return &Node{
		Info:     i,
		path:     path,
		children: children,
	}
}

type Node struct {
	Info     *ResolutionInfo
	path     tree.Path
	children map[string]tree.Node
}

func (n *Node) Path() tree.Path {
	return n.path
}

func (n *Node) Value() interface{} {
	return n.Info
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

func (n *Node) encoder() (*nodeEncoder, error) {
	children := make(map[string]nodeEncoder)
	for name, child := range n.children {
		c, ok := child.(*Node)
		if !ok {
			return nil, fmt.Errorf("node %v child %v is not a reference node", n.path.String(), name)
		}

		e, err := c.encoder()
		if err != nil {
			return nil, err
		}

		children[name] = *e
	}

	e := &nodeEncoder{
		Info:     n.Info,
		Children: children,
	}
	return e, nil
}

func nodeFromEncoder(e *nodeEncoder, path tree.Path) *Node {
	children := make(map[string]tree.Node)
	for name, child := range e.Children {

		children[name] = nodeFromEncoder(&child, path.Child(name))
	}

	return &Node{
		Info:     e.Info,
		path:     path,
		children: children,
	}
}

type nodeEncoder struct {
	Info     *ResolutionInfo        `json:"info"`
	Children map[string]nodeEncoder `json:"children"`
}

func (n *Node) MarshalJSON() ([]byte, error) {
	e, err := n.encoder()
	if err != nil {
		return nil, err
	}

	return json.Marshal(&e)
}

func (n *Node) UnmarshalJSON(data []byte) error {
	var e nodeEncoder
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	*n = *nodeFromEncoder(&e, tree.RootPath())
	return nil
}
