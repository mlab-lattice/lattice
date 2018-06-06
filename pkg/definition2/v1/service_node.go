package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition2/resource"
	"github.com/mlab-lattice/lattice/pkg/definition2/tree"
)

type ServiceNode struct {
	parent  tree.Node
	path    tree.NodePath
	service *Service
}

func NewServiceNode(service *Service, name string, parent tree.Node) *ServiceNode {
	return &ServiceNode{
		parent:  parent,
		path:    tree.ChildNodePath(parent.Path(), name),
		service: service,
	}
}

func (n *ServiceNode) Parent() tree.Node {
	return n.parent
}

func (n *ServiceNode) Path() tree.NodePath {
	return n.path
}

func (n *ServiceNode) Resource() resource.Interface {
	return n.service
}

func (n *ServiceNode) Service() *Service {
	return n.service
}
