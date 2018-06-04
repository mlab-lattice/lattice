package tree

import (
	"github.com/mlab-lattice/lattice/pkg/definition"
)

type ServiceNodePath NodePath

type ServiceNode struct {
	parent     Node
	path       NodePath
	definition *definition.Service
}

func NewServiceNode(def *definition.Service, parent Node) (*ServiceNode, error) {
	s := &ServiceNode{
		parent:     parent,
		path:       getPath(parent, def.Name),
		definition: def,
	}
	return s, nil
}

func (s *ServiceNode) Parent() Node {
	return s.parent
}

func (s *ServiceNode) Path() NodePath {
	return s.path
}

func (s *ServiceNode) Subsystems() map[NodePath]Node {
	return nil
}

func (s *ServiceNode) Definition() interface{} {
	return s.definition
}
