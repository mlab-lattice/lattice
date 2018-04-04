package tree

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition"
)

type SystemNode struct {
	parent         Node
	path           NodePath
	subsystemNodes map[NodePath]Node
	definition     definition.System
}

func NewSystemNode(definition definition.System, parent Node) (*SystemNode, error) {
	s := &SystemNode{
		parent:         parent,
		path:           getPath(parent, definition),
		definition:     definition,
		subsystemNodes: map[NodePath]Node{},
	}

	for _, subsystem := range definition.Subsystems() {
		child, err := NewNode(subsystem, Node(s))
		if err != nil {
			return nil, err
		}

		// Add child Node to subsystem
		childPath := child.Path()
		if _, exists := s.subsystemNodes[childPath]; exists {
			return nil, fmt.Errorf("System has multiple subsystems named %v", childPath)
		}

		s.subsystemNodes[childPath] = child
	}

	return s, nil
}

func (s *SystemNode) Type() string {
	return s.definition.Type()
}

func (s *SystemNode) Name() string {
	return s.definition.Name()
}

func (s *SystemNode) Description() string {
	return s.definition.Description()
}

func (s *SystemNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.definition)
}

func (s *SystemNode) Parent() Node {
	return s.parent
}

func (s *SystemNode) Path() NodePath {
	return NodePath(s.path)
}

func (s *SystemNode) Definition() definition.Interface {
	return s.definition
}

func (s *SystemNode) Subsystems() map[NodePath]Node {
	return s.subsystemNodes
}

func (s *SystemNode) Services() map[NodePath]*ServiceNode {
	svcNodes := map[NodePath]*ServiceNode{}

	for _, subsystem := range s.Subsystems() {
		for path, svcNode := range subsystem.Services() {
			svcNodes[path] = svcNode
		}
	}

	return svcNodes
}
