package tree

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/block"
)

type SystemNodePath NodePath

type SystemNode struct {
	parent     Node
	path       NodePath
	subsystems map[NodePath]Node
	definition *definition.System
}

func NewSystemNode(def *definition.System, parent Node) (*SystemNode, error) {
	s := &SystemNode{
		parent:     parent,
		path:       getPath(parent, def.Name),
		subsystems: make(map[NodePath]Node),
		definition: def,
	}

	for _, subsystem := range def.Subsystems {
		child, err := NewNode(subsystem, s)
		if err != nil {
			return nil, err
		}

		// Add child Node to subsystem
		childPath := child.Path()
		if _, exists := s.subsystems[childPath]; exists {
			return nil, fmt.Errorf("system has multiple subsystems named %v", childPath)
		}

		s.subsystems[childPath] = child
	}

	return s, nil
}

func (s *SystemNode) Parent() Node {
	return s.parent
}

func (s *SystemNode) Path() NodePath {
	return s.path
}

func (s *SystemNode) Subsystems() map[NodePath]Node {
	return s.subsystems
}

func (s *SystemNode) Definition() interface{} {
	return s.definition
}

func (s *SystemNode) Services() map[NodePath]*ServiceNode {
	services := make(map[NodePath]*ServiceNode)

	for _, subsystem := range s.Subsystems() {
		switch s := subsystem.(type) {
		case *SystemNode:
			for path, service := range s.Services() {
				services[path] = service
			}

		case *ServiceNode:
			services[s.Path()] = s
		}
	}

	return services
}

func (s *SystemNode) NodePools() map[string]block.NodePool {
	return s.definition.NodePools
}

func (s *SystemNode) UnmarshalJSON(data []byte) error {
	var def *definition.System
	err := json.Unmarshal(data, &def)
	if err != nil {
		return err
	}

	// unmarshalling a SystemNode will set it to be the root
	n, err := NewSystemNode(def, nil)
	if err != nil {
		return err
	}

	*s = *n
	return nil
}
