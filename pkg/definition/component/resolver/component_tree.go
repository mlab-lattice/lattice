package resolver

import (
	"encoding/json"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type (
	// ComponentTreeWalkFn is the function type invoked during a component tree walk.
	ComponentTreeWalkFn func(tree.Path, *ResolutionInfo) tree.WalkContinuation
	// V1TreeJobWalkFn is the function type invoked during a v1 job walk.
	V1TreeJobWalkFn func(tree.Path, *definitionv1.Job, *ResolutionInfo) tree.WalkContinuation
	// V1TreeNodePoolWalkFn is the function type invoked during a v1 node pool walk.
	V1TreeNodePoolWalkFn func(tree.PathSubcomponent, *definitionv1.NodePool) tree.WalkContinuation
	// V1TreeReferenceWalkFn is the function type invoked during a v1 reference walk.
	V1TreeReferenceWalkFn func(tree.Path, *definitionv1.Reference, *ResolutionInfo) tree.WalkContinuation
	// V1TreeServiceWalkFn is the function type invoked during a v1 service walk.
	V1TreeServiceWalkFn func(tree.Path, *definitionv1.Service, *ResolutionInfo) tree.WalkContinuation
	// V1TreeSystemWalkFn is the function type invoked during a v1 system walk.
	V1TreeSystemWalkFn func(tree.Path, *definitionv1.System, *ResolutionInfo) tree.WalkContinuation
	// V1TreeWorkloadWalkFn is the function type invoked during a v1 workload walk.
	V1TreeWorkloadWalkFn func(tree.Path, definitionv1.Workload, *ResolutionInfo) tree.WalkContinuation
)

// NewComponentTree returns an initialized ComponentTree.
func NewComponentTree() *ComponentTree {
	t := &ComponentTree{
		inner: tree.NewJSONRadix(
			func(i interface{}) (json.RawMessage, error) {
				return json.Marshal(&i)
			},
			func(data json.RawMessage) (interface{}, error) {
				var info ResolutionInfo
				if err := json.Unmarshal(data, &info); err != nil {
					return nil, err
				}

				return &info, nil
			},
		),
	}
	t.v1 = &V1Tree{ComponentTree: t}
	return t
}

// ComponentTree provides efficient path-based access to info about the resolution of a given
// tree of components.
type ComponentTree struct {
	inner *tree.JSONRadix
	v1    *V1Tree
}

// Insert adds component resolution information about a path.
func (t *ComponentTree) Insert(p tree.Path, c *ResolutionInfo) (*ResolutionInfo, bool) {
	prev, replaced := t.inner.Insert(p, c)
	if !replaced {
		return nil, false
	}

	return prev.(*ResolutionInfo), true
}

// Get retrieves component resolution information about a path.
func (t *ComponentTree) Get(p tree.Path) (*ResolutionInfo, bool) {
	c, ok := t.inner.Get(p)
	if !ok {
		return nil, false
	}

	return c.(*ResolutionInfo), true
}

// Delete removes component resolution information about a path.
func (t *ComponentTree) Delete(p tree.Path) (*ResolutionInfo, bool) {
	c, ok := t.inner.Delete(p)
	if !ok {
		return nil, false
	}

	return c.(*ResolutionInfo), true
}

// Len returns the number of elements in the tree.
func (t *ComponentTree) Len() int {
	return t.inner.Len()
}

// ReplacePrefix removes existing component resolution information about a path
// and all paths below it, and replaces it with the information from the supplied
// ComponentTree.
func (t *ComponentTree) ReplacePrefix(p tree.Path, other *ComponentTree) {
	t.inner.ReplacePrefix(p, other.inner.Radix)
}

// Walk walks the Component tree, invoking the supplied function on each path.
func (t *ComponentTree) Walk(fn ComponentTreeWalkFn) {
	t.inner.Walk(func(path tree.Path, i interface{}) tree.WalkContinuation {
		return fn(path, i.(*ResolutionInfo))
	})
}

// V1 returns a V1 tree allowing retrieval of v1 components in the tree.
func (t *ComponentTree) V1() *V1Tree {
	return t.v1
}

// MarshalJSON fulfills the json.Marshaller interface.
func (t *ComponentTree) MarshalJSON() ([]byte, error) {
	return json.Marshal(&t.inner)
}

// MarshalJSON fulfills the json.Unmarshaller interface.
func (t *ComponentTree) UnmarshalJSON(data []byte) error {
	t2 := NewComponentTree()
	if err := json.Unmarshal(data, &t2.inner); err != nil {
		return err
	}

	*t = *t2
	return nil
}

// V1Tree provides an overlay on top of a Component tree to access v1 components in the tree.
type V1Tree struct {
	*ComponentTree
}

// Jobs walks the Component tree, invoking the supplied function on each path that contains a v1/job.
func (t *V1Tree) Jobs(fn V1TreeJobWalkFn) {
	t.ComponentTree.Walk(func(path tree.Path, i *ResolutionInfo) tree.WalkContinuation {
		job, ok := i.Component.(*definitionv1.Job)
		if !ok {
			return tree.ContinueWalk
		}

		return fn(path, job, i)
	})
}

// NodePools walks the Component tree, invoking the supplied function on each path that contains a v1/node-pool.
func (t *V1Tree) NodePools(fn V1TreeNodePoolWalkFn) {
	t.Systems(func(path tree.Path, system *definitionv1.System, info *ResolutionInfo) tree.WalkContinuation {
		for name, nodePool := range system.NodePools {
			// FIXME(kevindrosendahl): what to do in the event of an empty string node pool?
			subcomponent, _ := tree.NewPathSubcomponentFromParts(path, name)
			if !fn(subcomponent, &nodePool) {
				return tree.HaltWalk
			}
		}

		return tree.ContinueWalk
	})
}

// References walks the Component tree, invoking the supplied function on each path that contains a v1/reference.
func (t *V1Tree) References(fn V1TreeReferenceWalkFn) {
	t.ComponentTree.Walk(func(path tree.Path, i *ResolutionInfo) tree.WalkContinuation {
		reference, ok := i.Component.(*definitionv1.Reference)
		if !ok {
			return tree.ContinueWalk
		}

		return fn(path, reference, i)
	})
}

// Services walks the Component tree, invoking the supplied function on each path that contains a v1/services.
func (t *V1Tree) Services(fn V1TreeServiceWalkFn) {
	t.ComponentTree.Walk(func(path tree.Path, i *ResolutionInfo) tree.WalkContinuation {
		service, ok := i.Component.(*definitionv1.Service)
		if !ok {
			return tree.ContinueWalk
		}

		return fn(path, service, i)
	})
}

// Systems walks the Component tree, invoking the supplied function on each path that contains a v1/system.
func (t *V1Tree) Systems(fn V1TreeSystemWalkFn) {
	t.ComponentTree.Walk(func(path tree.Path, i *ResolutionInfo) tree.WalkContinuation {
		system, ok := i.Component.(*definitionv1.System)
		if !ok {
			return tree.ContinueWalk
		}

		return fn(path, system, i)
	})
}

// Workloads walks the Component tree, invoking the supplied function on each path whose
// component fulfills the Workload interface.
func (t *V1Tree) Workloads(fn V1TreeWorkloadWalkFn) {
	t.ComponentTree.Walk(func(path tree.Path, i *ResolutionInfo) tree.WalkContinuation {
		workload, ok := i.Component.(definitionv1.Workload)
		if !ok {
			return tree.ContinueWalk
		}

		return fn(path, workload, i)
	})
}
