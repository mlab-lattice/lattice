package resolver

import (
	"encoding/json"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type (
	ComponentTreeWalkFn   func(tree.Path, *ResolutionInfo) bool
	V1TreeJobWalkFn       func(tree.Path, *definitionv1.Job, *ResolutionInfo) bool
	V1TreeNodePoolWalkFn  func(tree.PathSubcomponent, *definitionv1.NodePool) bool
	V1TreeReferenceWalkFn func(tree.Path, *definitionv1.Reference, *ResolutionInfo) bool
	V1TreeServiceWalkFn   func(tree.Path, *definitionv1.Service, *ResolutionInfo) bool
	V1TreeSystemWalkFn    func(tree.Path, *definitionv1.System, *ResolutionInfo) bool
	V1TreeWorkloadWalkFn  func(tree.Path, definitionv1.Workload, *ResolutionInfo) bool
)

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

type ComponentTree struct {
	inner *tree.JSONRadix
	v1    *V1Tree
}

func (t *ComponentTree) Insert(p tree.Path, c *ResolutionInfo) (*ResolutionInfo, bool) {
	prev, replaced := t.inner.Insert(p, c)
	if !replaced {
		return nil, false
	}

	return prev.(*ResolutionInfo), true
}

func (t *ComponentTree) Get(p tree.Path) (*ResolutionInfo, bool) {
	c, ok := t.inner.Get(p)
	if !ok {
		return nil, false
	}

	return c.(*ResolutionInfo), true
}

func (t *ComponentTree) Delete(p tree.Path) (*ResolutionInfo, bool) {
	c, ok := t.inner.Delete(p)
	if !ok {
		return nil, false
	}

	return c.(*ResolutionInfo), true
}

func (t *ComponentTree) ReplacePrefix(p tree.Path, other *ComponentTree) {
	t.inner.ReplacePrefix(p, other.inner.Radix)
}

func (t *ComponentTree) Walk(fn ComponentTreeWalkFn) {
	t.inner.Walk(func(path tree.Path, i interface{}) bool {
		return fn(path, i.(*ResolutionInfo))
	})
}

func (t *ComponentTree) V1() *V1Tree {
	return t.v1
}

func (t *ComponentTree) MarshalJSON() ([]byte, error) {
	return json.Marshal(&t.inner)
}

func (t *ComponentTree) UnmarshalJSON(data []byte) error {
	t2 := NewComponentTree()
	if err := json.Unmarshal(data, &t2.inner); err != nil {
		return err
	}

	*t = *t2
	return nil
}

type V1Tree struct {
	*ComponentTree
}

func (t *V1Tree) Jobs(fn V1TreeJobWalkFn) {
	t.ComponentTree.Walk(func(path tree.Path, i *ResolutionInfo) bool {
		job, ok := i.Component.(*definitionv1.Job)
		if !ok {
			return false
		}

		return fn(path, job, i)
	})
}

func (t *V1Tree) NodePools(fn V1TreeNodePoolWalkFn) {
	t.Systems(func(path tree.Path, system *definitionv1.System, info *ResolutionInfo) bool {
		for name, nodePool := range system.NodePools {
			// FIXME(kevindrosendahl): what to do in the event of an empty string node pool?
			subcomponent, _ := tree.NewPathSubcomponentFromParts(path, name)
			if !fn(subcomponent, &nodePool) {
				return false
			}
		}

		return true
	})
}

func (t *V1Tree) References(fn V1TreeReferenceWalkFn) {
	t.ComponentTree.Walk(func(path tree.Path, i *ResolutionInfo) bool {
		reference, ok := i.Component.(*definitionv1.Reference)
		if !ok {
			return false
		}

		return fn(path, reference, i)
	})
}

func (t *V1Tree) Services(fn V1TreeServiceWalkFn) {
	t.ComponentTree.Walk(func(path tree.Path, i *ResolutionInfo) bool {
		service, ok := i.Component.(*definitionv1.Service)
		if !ok {
			return false
		}

		return fn(path, service, i)
	})
}

func (t *V1Tree) Systems(fn V1TreeSystemWalkFn) {
	t.ComponentTree.Walk(func(path tree.Path, i *ResolutionInfo) bool {
		system, ok := i.Component.(*definitionv1.System)
		if !ok {
			return false
		}

		return fn(path, system, i)
	})
}

func (t *V1Tree) Workloads(fn V1TreeWorkloadWalkFn) {
	t.ComponentTree.Walk(func(path tree.Path, i *ResolutionInfo) bool {
		workload, ok := i.Component.(definitionv1.Workload)
		if !ok {
			return false
		}

		return fn(path, workload, i)
	})
}
