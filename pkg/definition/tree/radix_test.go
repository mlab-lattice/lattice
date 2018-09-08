package tree

import (
	"encoding/json"
	"testing"
)

func TestRadix(t *testing.T) {
	r := NewRadix()
	m := seedShallow(r)

	for p, v := range m {
		i, ok := r.Get(p)
		if !ok {
			t.Fatalf("expected path %v to exist", p.String())
		}

		if i != v {
			t.Fatalf("expected path %v to contain %v but contains %v", p.String(), v, i)
		}
	}
}

func TestRadix_Delete(t *testing.T) {
	r := NewRadix()
	seedShallow(r)

	p := RootPath().Child("e")
	v, ok := r.Delete(p)
	if !ok {
		t.Fatalf("expected to be able to remove %v", p.String())
	}

	if v != "e" {
		t.Errorf("got unexpected value for removed path: %v", v)
	}

	_, ok = r.Get(p)
	if ok {
		t.Errorf("expected to not be able to retrieve removed path")
	}
}

func TestRadix_DeletePrefix(t *testing.T) {
	r := NewRadix()
	m := seedMixed(r)

	prefix := RootPath().Child("a").Child("b").Child("c")
	expectedRemovals := 26 - 2
	removals := r.DeletePrefix(prefix)
	if removals != expectedRemovals {
		t.Fatalf("expected to remove %v entries, removed %v", expectedRemovals, removals)
	}

	expectedLen := len(m) - removals
	if r.Len() != expectedLen {
		t.Fatalf("expected %v remaining entries, found %v", expectedLen, r.Len())
	}

	r.Walk(func(p Path, v interface{}) bool {
		if p.HasPrefix(prefix) {
			t.Errorf("expected to not find any paths with prefix %v, found %v", prefix.String(), p.String())
		}

		return false
	})
}

func TestRadix_Walk(t *testing.T) {
	r := NewRadix()
	seedMixed(r)
	last := RootPath()
	first := true
	r.Walk(func(p Path, v interface{}) bool {
		if first {
			last = p
			first = false
			return false
		}

		longer := last
		shorter := p
		if p.Depth() > longer.Depth() {
			longer = p
			shorter = last
		}

		longerPrefix, _ := longer.Prefix(shorter.Depth())
		if longerPrefix > shorter {
			t.Fatalf("expected walk to preserve order but it did not (%v came before %v)", last.String(), p.String())
		}

		last = p
		return false
	})
}

func TestRadix_WalkPrefix(t *testing.T) {
	r := NewRadix()
	seedMixed(r)
	prefix := RootPath().Child("a").Child("b").Child("c")
	r.WalkPrefix(prefix, func(p Path, v interface{}) bool {
		if !p.HasPrefix(prefix) {
			t.Fatalf("expected prefix walk path %v to have prefix %v", p.String(), prefix.String())
		}
		return false
	})
}

func TestJSONRadix(t *testing.T) {
	type s struct {
		Foo int `json:"foo1"`
		Bar int `json:"foo2"`
	}
	r1 := NewJSONRadix(
		func(v interface{}) (json.RawMessage, error) {
			return json.Marshal(v)
		},
		nil,
	)

	vals := []struct {
		p Path
		v s
	}{
		{
			p: RootPath().Child("a").Child("b"),
			v: s{Foo: 1, Bar: 2},
		},
		{
			p: RootPath().Child("b"),
			v: s{Foo: 2, Bar: 3},
		},
	}

	for _, v := range vals {
		r1.Insert(v.p, v.v)
	}

	r2 := NewJSONRadix(
		nil,
		func(data json.RawMessage) (interface{}, error) {
			var s2 s
			err := json.Unmarshal(data, &s2)
			return s2, err
		},
	)

	data, err := json.Marshal(&r1)
	if err != nil {
		t.Fatalf("error marshalling radix: %v", err)
	}

	if err := json.Unmarshal(data, &r2); err != nil {
		t.Fatalf("error unmarshalling radix: %v", err)
	}

	if r1.Len() != r2.Len() {
		t.Fatalf("expected marshalled radix to contain %v elements but it contains %v", r1.Len(), r2.Len())
	}

	r1.Walk(func(p Path, i interface{}) bool {
		v, ok := r2.Get(p)
		if !ok {
			t.Fatalf("expected marshalled radix to contain %v but it does not", p.String())
		}

		if i != v {
			t.Fatalf("expected marshalled radix path %v to be %v but is %v", p.String(), i, v)
		}

		return false
	})
}

func seedShallow(r *Radix) map[Path]interface{} {
	m := make(map[Path]interface{})
	for i := 0; i < 26; i++ {
		c := rune('a' + i)
		p := RootPath().Child(string(c))

		v := p.ToDomain()
		m[p] = v
		r.Insert(p, v)
	}

	return m
}

func seedDeep(r *Radix) map[Path]interface{} {
	m := make(map[Path]interface{})
	for i := 0; i < 26; i++ {
		p := RootPath()
		for j := 0; j <= i; j++ {
			c := rune('a' + j)
			p = p.Child(string(c))
		}

		v := p.ToDomain()
		m[p] = v
		r.Insert(p, v)
	}

	return m
}

func seedMixed(r *Radix) map[Path]interface{} {
	m := seedShallow(r)
	m2 := seedDeep(r)
	for k, v := range m2 {
		m[k] = v
	}

	return m
}
