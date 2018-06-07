package v1

import (
	"fmt"
	"github.com/mlab-lattice/lattice/pkg/definition/resource"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

func NewNode(r resource.Interface, name string, parent tree.Node) (tree.Node, error) {
	switch res := r.(type) {
	case *Job:
		return NewJobNode(res, name, parent), nil

	case *Service:
		return NewServiceNode(res, name, parent), nil

	case *System:
		return NewSystemNode(res, name, parent)

	default:
		return nil, fmt.Errorf("unrecognized resource type: %v", r.Type().String())
	}
}
