package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func NewNode(c component.Interface, name string, parent tree.Node) (tree.Node, error) {
	switch res := c.(type) {
	case *Job:
		return NewJobNode(name, parent, res), nil

	case *Reference:
		return NewReferenceNode(name, parent, res), nil

	case *Service:
		return NewServiceNode(name, parent, res), nil

	case *System:
		return NewSystemNode(name, parent, res)

	default:
		return nil, fmt.Errorf("unrecognized component type: %v", c.Type().String())
	}
}
