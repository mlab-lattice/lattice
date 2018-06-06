package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition2/resource"
	"github.com/mlab-lattice/lattice/pkg/definition2/tree"
)

type JobNode struct {
	parent tree.Node
	path   tree.NodePath
	job    *Job
}

func NewJobNode(job *Job, name string, parent tree.Node) *JobNode {
	return &JobNode{
		parent: parent,
		path:   tree.ChildNodePath(parent.Path(), name),
		job:    job,
	}
}

func (n *JobNode) Parent() tree.Node {
	return n.parent
}

func (n *JobNode) Path() tree.NodePath {
	return n.path
}

func (n *JobNode) Resource() resource.Interface {
	return n.job
}

func (n *JobNode) Job() *Job {
	return n.job
}
