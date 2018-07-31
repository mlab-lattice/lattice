package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type ServiceNode struct {
	parent  tree.ComponentNode
	path    tree.NodePath
	service *Service
}

func NewServiceNode(service *Service, name string, parent tree.ComponentNode) *ServiceNode {
	return &ServiceNode{
		parent:  parent,
		path:    parent.Path().Child(name),
		service: service,
	}
}

func (n *ServiceNode) Parent() tree.ComponentNode {
	return n.parent
}

func (n *ServiceNode) Path() tree.NodePath {
	return n.path
}

func (n *ServiceNode) Value() interface{} {
	return n.service
}

func (n *ServiceNode) Component() component.Interface {
	return n.service
}

func (n *ServiceNode) Service() *Service {
	return n.service
}
