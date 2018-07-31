package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type JobNode struct {
	parent tree.ComponentNode
	path   tree.NodePath
	job    *Job
}

func NewJobNode(job *Job, name string, parent tree.ComponentNode) *JobNode {
	return &JobNode{
		parent: parent,
		path:   parent.Path().Child(name),
		job:    job,
	}
}

func (n *JobNode) Parent() tree.ComponentNode {
	return n.parent
}

func (n *JobNode) Path() tree.NodePath {
	return n.path
}

func (n *JobNode) Component() component.Interface {
	return n.job
}

func (n *JobNode) Value() interface{} {
	return n.job
}

func (n *JobNode) Job() *Job {
	return n.job
}
