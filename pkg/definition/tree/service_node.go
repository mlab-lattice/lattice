package tree

import (
	"encoding/json"

	"github.com/mlab-lattice/system/pkg/definition"
)

type ServiceNode struct {
	parent     Node
	path       NodePath
	definition *definition.Service
}

func NewServiceNode(definition *definition.Service, parent Node) (*ServiceNode, error) {
	if err := definition.Validate(nil); err != nil {
		return nil, err
	}

	s := &ServiceNode{
		parent:     parent,
		path:       getPath(parent, definition),
		definition: definition,
	}
	return s, nil
}

// Implement the encoding/json.Marshaller interface.
func (s *ServiceNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.definition)
}

// Implement the Node interface.
func (s *ServiceNode) Parent() Node {
	return s.parent
}

func (s *ServiceNode) Path() NodePath {
	return NodePath(s.path)
}

func (s *ServiceNode) Name() string {
	return s.definition.Meta.Name
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
