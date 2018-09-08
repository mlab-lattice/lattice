package v1

import (
	"encoding/json"
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type (
	TreeWalkFn          func(tree.Path, component.Interface) bool
	TreeJobWalkFn       func(tree.Path, *Job) bool
	TreeReferenceWalkFn func(tree.Path, *Reference) bool
	TreeServiceWalkFn   func(tree.Path, *Service) bool
	TreeWorkloadWalkFn  func(tree.Path, Workload) bool
)

func NewTree() *Tree {
	return &Tree{
		inner: tree.NewJSONRadix(
			func(i interface{}) (json.RawMessage, error) {
				return json.Marshal(&i)
			},
			func(data json.RawMessage) (interface{}, error) {
				return NewComponentFromJSON(data)
			},
		),
	}
}

type Tree struct {
	inner *tree.JSONRadix
}

func (t *Tree) Insert(p tree.Path, c component.Interface) (component.Interface, bool) {
	prev, replaced := t.inner.Insert(p, c)
	if !replaced {
		return nil, false
	}

	return prev.(component.Interface), true
}

func (t *Tree) Get(p tree.Path) (component.Interface, bool) {
	c, ok := t.inner.Get(p)
	if !ok {
		return nil, false
	}

	return c.(component.Interface), true
}

func (t *Tree) Delete(p tree.Path) (component.Interface, bool) {
	c, ok := t.inner.Delete(p)
	if !ok {
		return nil, false
	}

	return c.(component.Interface), true
}

func (t *Tree) DeletePrefix(p tree.Path) int {
	return t.inner.DeletePrefix(p)
}

func (t *Tree) Len() int {
	return t.inner.Len()
}

func (t *Tree) ReplacePrefix(p tree.Path, other *Tree) {
	t.inner.ReplacePrefix(p, t.inner.Radix)
}

func (t *Tree) Walk(fn TreeWalkFn) {
	t.inner.Walk(func(path tree.Path, i interface{}) bool {
		return fn(path, i.(component.Interface))
	})
}

func (t *Tree) Jobs(fn TreeJobWalkFn) {
	t.inner.Walk(func(path tree.Path, i interface{}) bool {
		job, ok := i.(*Job)
		if !ok {
			return false
		}

		return fn(path, job)
	})
}

func (t *Tree) References(fn TreeReferenceWalkFn) {
	t.inner.Walk(func(path tree.Path, i interface{}) bool {
		reference, ok := i.(*Reference)
		if !ok {
			return false
		}

		return fn(path, reference)
	})
}

func (t *Tree) Services(fn TreeServiceWalkFn) {
	t.inner.Walk(func(path tree.Path, i interface{}) bool {
		service, ok := i.(*Service)
		if !ok {
			return false
		}

		return fn(path, service)
	})
}

func (t *Tree) Workloads(fn TreeWorkloadWalkFn) {
	t.inner.Walk(func(path tree.Path, i interface{}) bool {
		workload, ok := i.(Workload)
		if !ok {
			return false
		}

		return fn(path, workload)
	})
}
