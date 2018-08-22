package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type JobNode struct {
	parent tree.Node
	path   tree.Path
	job    *Job
}

func NewJobNode(name string, parent tree.Node, job *Job) *JobNode {
	return &JobNode{
		parent: parent,
		path:   parent.Path().Child(name),
		job:    job,
	}
}

func (n *JobNode) Path() tree.Path {
	return n.path
}

func (n *JobNode) Value() interface{} {
	return n.job
}

func (n *JobNode) Parent() tree.Node {
	return n.parent
}

func (n *JobNode) Children() map[string]tree.Node {
	return nil
}

func (n *JobNode) Component() component.Interface {
	return n.job
}

func (n *JobNode) Job() *Job {
	return n.job
}
