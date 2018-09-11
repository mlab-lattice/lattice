package tree

import (
	"encoding/json"

	"github.com/armon/go-radix"
)

type (
	JSONRadixMarshalFn   func(interface{}) (json.RawMessage, error)
	JSONRadixUnmarshalFn func(json.RawMessage) (interface{}, error)
	RadixWalkFn          func(Path, interface{}) WalkContinuation
	WalkContinuation     bool
)

const (
	ContinueWalk WalkContinuation = false
	HaltWalk     WalkContinuation = true
)

// NewRadix returns a radix tree.
func NewRadix() *Radix {
	return &Radix{inner: radix.New()}
}

// Radix provides efficient insertion, retrieval, and deletion,
// as well as prefixed retrieval and deletion on paths.
type Radix struct {
	inner *radix.Tree
}

// Insert inserts a value for the given path.
func (r *Radix) Insert(p Path, v interface{}) (interface{}, bool) {
	return r.inner.Insert(p.String(), v)
}

// Get retrieves the value for a given path if it exists, and a bool
// indicating if the path exists.
func (r *Radix) Get(p Path) (interface{}, bool) {
	return r.inner.Get(p.String())
}

// Delete deletes a path from the tree if it exists, and returns
// the deleted value if it exists as well as a boolean indicating whether
// the path existed.
func (r *Radix) Delete(p Path) (interface{}, bool) {
	return r.inner.Delete(p.String())
}

// Delete prefix removes all paths that are prefixed by the given prefix
// (including the given prefix) from the tree.
func (r *Radix) DeletePrefix(p Path) int {
	return r.inner.DeletePrefix(p.String())
}

// Len returns the number of paths in the tree.
func (r *Radix) Len() int {
	return r.inner.Len()
}

// ReplacePrefix removes all entries under the supplied prefix in the radix tree
// and replaces them with the entries under the prefix from the other supplied
// radix tree.
func (r *Radix) ReplacePrefix(p Path, other *Radix) {
	r.DeletePrefix(p)
	other.WalkPrefix(p, func(path Path, i interface{}) WalkContinuation {
		r.Insert(path, i)
		return ContinueWalk
	})
}

// Walk walks the tree in lexical order, invoking the supplied function
// at each node.
func (r *Radix) Walk(fn RadixWalkFn) {
	r.inner.Walk(walkFn(fn))
}

// WalkPrefix walks all paths including and under the prefix in the tree.
func (r *Radix) WalkPrefix(p Path, fn RadixWalkFn) {
	r.inner.WalkPrefix(p.String(), walkFn(fn))
}

func walkFn(fn RadixWalkFn) radix.WalkFn {
	return func(s string, v interface{}) bool {
		p, _ := NewPath(s)
		return bool(fn(p, v))
	}
}

// NewJSONRadix creates a new radix with the supplied marshaller and unmarshaller functions
// and returns it.
func NewJSONRadix(marshaller JSONRadixMarshalFn, unmarshaller JSONRadixUnmarshalFn) *JSONRadix {
	return &JSONRadix{
		Radix:        NewRadix(),
		marshaller:   marshaller,
		unmarshaller: unmarshaller,
	}
}

// JSONRadix is a Radix tree that is capable of being (de)serialized to/from
// JSON using the supplied marshalling/unmarshalling functions.
type JSONRadix struct {
	*Radix
	marshaller   JSONRadixMarshalFn
	unmarshaller JSONRadixUnmarshalFn
}

// MarshalJSON fulfills the json.Marshaller interface.
func (r *JSONRadix) MarshalJSON() ([]byte, error) {
	out := make(map[Path]json.RawMessage, r.inner.Len())
	var err error
	r.Walk(func(p Path, v interface{}) WalkContinuation {
		var data json.RawMessage
		data, err = r.marshaller(v)
		if err != nil {
			return HaltWalk
		}

		out[p] = data
		return ContinueWalk
	})

	if err != nil {
		return nil, err
	}

	return json.Marshal(&out)
}

// MarshalJSON fulfills the json.Unmarshaller interface.
func (r *JSONRadix) UnmarshalJSON(data []byte) error {
	in := make(map[Path]json.RawMessage)
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}

	r.Radix = NewRadix()
	for p, v := range in {
		i, err := r.unmarshaller(v)
		if err != nil {
			return err
		}

		r.Insert(p, i)
	}

	return nil
}
