package tree

import (
	"encoding/json"

	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/block"
)

type ServiceNode struct {
	parent     Node
	path       NodePath
	definition definition.Service
}

func NewServiceNode(definition definition.Service, parent Node) (*ServiceNode, error) {
	s := &ServiceNode{
		parent:     parent,
		path:       getPath(parent, definition),
		definition: definition,
	}
	return s, nil
}

func (s *ServiceNode) Type() string {
	return s.definition.Type()
}

func (s *ServiceNode) Name() string {
	return s.definition.Name()
}

func (s *ServiceNode) Description() string {
	return s.definition.Description()
}

func (s *ServiceNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.definition)
}

func (s *ServiceNode) Parent() Node {
	return s.parent
}

func (s *ServiceNode) Path() NodePath {
	return NodePath(s.path)
}

func (s *ServiceNode) Definition() definition.Interface {
	return definition.Interface(s.definition)
}

func (s *ServiceNode) Subsystems() map[NodePath]Node {
	return map[NodePath]Node{}
}

func (s *ServiceNode) Services() map[NodePath]*ServiceNode {
	return map[NodePath]*ServiceNode{
		s.Path(): s,
	}
}

func (s *ServiceNode) NodePools() map[string]block.NodePool {
	return nil
}
