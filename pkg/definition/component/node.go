package component

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type Node interface {
	tree.Node
	Component() Interface
}

func Lookup(n Node, p tree.Path) (Node, bool, error) {
	r, ok, err := tree.Lookup(n, p)
	if err != nil {
		return nil, false, err
	}

	if !ok {
		return nil, false, nil
	}

	cn, ok := r.(Node)
	if !ok {
		return nil, false, fmt.Errorf("component node child %v was not a component node", p.String())
	}

	return cn, true, nil
}
