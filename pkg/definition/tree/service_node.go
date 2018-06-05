package tree

import (
	"encoding/json"

	"github.com/mlab-lattice/lattice/pkg/definition"
)

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

func (s *ServiceNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(&s.definition)
}

func (s *ServiceNode) UnmarshalJSON(data []byte) error {
	var def *definition.Service
	err := json.Unmarshal(data, &def)
	if err != nil {
		return err
	}

	// unmarshalling a ServiceNode will set it to be the root
	n, err := NewServiceNode(def, nil)
	if err != nil {
		return err
	}

	*s = *n
	return nil
}
