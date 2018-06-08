package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func NewNode(c component.Interface, name string, parent tree.Node) (tree.Node, error) {
	switch res := c.(type) {
	case *Job:
		return NewJobNode(res, name, parent), nil

	case *Service:
		return NewServiceNode(res, name, parent), nil

	case *System:
		return NewSystemNode(res, name, parent)

	default:
		return nil, fmt.Errorf("unrecognized component type: %v", c.Type().String())
	}
}
