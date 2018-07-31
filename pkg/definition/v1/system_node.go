package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type SystemNode struct {
	parent tree.ComponentNode
	path   tree.NodePath

	system *System

	components map[string]tree.ComponentNode
	systems    map[string]*SystemNode
	services   map[string]*ServiceNode
	jobs       map[string]*JobNode
	references map[string]*ReferenceNode
}

func NewSystemNode(system *System, name string, parent tree.ComponentNode) (*SystemNode, error) {
	path := tree.RootNodePath()
	if parent != nil {
		path = parent.Path().Child(name)
	}

	node := &SystemNode{
		parent: nil,
		path:   path,

		system: system,

		components: make(map[string]tree.ComponentNode),
		systems:    make(map[string]*SystemNode),
		services:   make(map[string]*ServiceNode),
		jobs:       make(map[string]*JobNode),
		references: make(map[string]*ReferenceNode),
	}

	for n, c := range system.Components {
		componentNode, err := NewNode(c, n, node)
		if err != nil {
			return nil, err
		}

		node.components[n] = componentNode

		switch typedNode := componentNode.(type) {
		case *JobNode:
			node.jobs[n] = typedNode

		case *ReferenceNode:
			node.references[n] = typedNode

		case *ServiceNode:
			node.services[n] = typedNode

		case *SystemNode:
			node.systems[n] = typedNode

		default:
			return nil, fmt.Errorf("unrecognized node type")
		}
	}

	return node, nil
}

func (n *SystemNode) Parent() tree.ComponentNode {
	return n.parent
}

func (n *SystemNode) Path() tree.NodePath {
	return n.path
}

func (n *SystemNode) Component() component.Interface {
	return n.system
}

func (n *SystemNode) System() *System {
	return n.system
}

func (n *SystemNode) NodePools() map[string]NodePool {
	return n.system.NodePools
}

func (n *SystemNode) Components() map[string]tree.ComponentNode {
	return n.components
}

func (n *SystemNode) Jobs() map[string]*JobNode {
	return n.jobs
}

// FIXME: come up with a better name for this
func (n *SystemNode) AllJobs() map[tree.NodePath]*JobNode {
	jobs := make(map[tree.NodePath]*JobNode)
	n.Walk(func(node *SystemNode) error {
		for _, job := range node.Jobs() {
			jobs[job.Path()] = job
		}

		return nil
	})

	return jobs
}

func (n *SystemNode) References() map[string]*ReferenceNode {
	return n.references
}

//func (n *SystemNode) ResolveReferences(r resolver.Interface) (*SystemNode, error) {
//	components := make(map[string]component.Interface)
//	for k, v := range n.Components() {
//		components[k] = v.Component()
//	}
//
//	for name, refNode := range n.References() {
//		ref := refNode.Reference()
//		switch {
//		case ref.File != nil:
//
//		}
//	}
//}

func (n *SystemNode) Services() map[string]*ServiceNode {
	return n.services
}

// FIXME: come up with a better name for this
func (n *SystemNode) AllServices() map[tree.NodePath]*ServiceNode {
	services := make(map[tree.NodePath]*ServiceNode)
	n.Walk(func(node *SystemNode) error {
		for _, service := range node.Services() {
			services[service.Path()] = service
		}

		return nil
	})

	return services
}

func (n *SystemNode) Systems() map[string]*SystemNode {
	return n.systems
}

func (n *SystemNode) Walk(fn func(*SystemNode) error) error {
	err := fn(n)
	if err != nil {
		return fmt.Errorf("error walking node %v: %v", n.Path().String(), err)
	}

	for _, subsystem := range n.Systems() {
		err := subsystem.Walk(fn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *SystemNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.system)
}

func (n *SystemNode) UnmarshalJSON(data []byte) error {
	var s *System
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	node, err := NewSystemNode(s, "", nil)
	if err != nil {
		return err
	}

	*n = *node
	return nil
}
