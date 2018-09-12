package tree

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNewPath(t *testing.T) {
	tests := []struct {
		d string
		p string
		e bool
		r Path
	}{
		{
			d: "empty string",
			p: "",
			e: true,
		},
		{
			d: "no leading slash",
			p: "foo/bar",
			e: true,
		},
		{
			d: "empty internal subpath",
			p: "/foo//bar",
			e: true,
		},
		{
			d: "empty trailing subpath",
			p: "/foo/bar/",
			e: true,
		},
		{
			d: "root",
			p: "/",
			e: false,
			r: Path("/"),
		},
		{
			d: "valid path",
			p: "/foo/Bar/BUZZ",
			e: false,
			r: Path("/foo/Bar/BUZZ"),
		},
	}

	for _, test := range tests {
		p, err := NewPath(test.p)
		if err != nil {
			if !test.e {
				t.Errorf("expected no error for %v but got %e", test.d, err)
			}
			continue
		}

		if test.e {
			t.Errorf("expected error for %v but got nil", test.d)
			continue
		}

		if !reflect.DeepEqual(p, test.r) {
			t.Errorf("expected %v but got %v for %v", p, test.r, test.d)
		}
	}
}

func TestPathFromDomain(t *testing.T) {
	tests := []struct {
		d string
		p string
		e bool
		r Path
	}{
		{
			d: "empty initial subdomain",
			p: ".bar.foo",
			e: true,
		},
		{
			d: "empty mid subdomain",
			p: "bar..foo",
			e: true,
		},
		{
			d: "empty trailing subdomain",
			p: "bar.foo.",
			e: true,
		},
		{
			d: "valid domain",
			p: "BUZZ.Bar.foo",
			e: false,
			r: Path("/foo/Bar/BUZZ"),
		},
	}

	for _, test := range tests {
		p, err := NewPathFromDomain(test.p)
		if err != nil {
			if !test.e {
				t.Errorf("expected no error for %v but got %e", test.d, err)
			}
			continue
		}

		if test.e {
			t.Errorf("expected error for %v but got nil", test.d)
			continue
		}

		if !reflect.DeepEqual(p, test.r) {
			t.Errorf("expected %v but got %v for %v", p, test.r, test.d)
		}
	}
}

func TestPath_ToDomain(t *testing.T) {
	p, err := NewPath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("expected no error for valid path but got %v", err)
	}

	domain := p.ToDomain()
	expectedDomain := "buzz.bar.foo"
	if domain != expectedDomain {
		t.Errorf("expected domain %v but got %v", expectedDomain, domain)
	}
}

func TestPath_Depth(t *testing.T) {
	p, err := NewPath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("expected no error for valid path but got %v", err)
	}

	depth := p.Depth()
	expectedDepth := 3
	if expectedDepth != depth {
		t.Errorf("expected depth %v but got %v", expectedDepth, depth)
	}

	p, err = NewPath("/foo/Bar")
	if err != nil {
		t.Errorf("expected no error for valid path but got %v", err)
	}

	depth = p.Depth()
	expectedDepth = 2
	if expectedDepth != depth {
		t.Errorf("expected depth %v but got %v", expectedDepth, depth)
	}

	p, err = NewPath("/foo")
	if err != nil {
		t.Errorf("expected no error for valid path but got %v", err)
	}

	depth = p.Depth()
	expectedDepth = 1
	if expectedDepth != depth {
		t.Errorf("expected depth %v but got %v", expectedDepth, depth)
	}
}

func TestPath_IsRoot(t *testing.T) {
	p, err := NewPath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("expected no error for valid path but got %v", err)
	}

	if p.IsRoot() {
		t.Errorf("expected Path to not be Root")
	}

	p, err = NewPath("/foo/Bar")
	if err != nil {
		t.Errorf("expected no error for valid path but got %v", err)
	}

	if p.IsRoot() {
		t.Errorf("expected Path to not be Root")
	}

	p, err = NewPath("/foo")
	if err != nil {
		t.Errorf("expected no error for valid path but got %v", err)
	}

	if p.IsRoot() {
		t.Errorf("expected Path to not be Root")
	}

	p, err = NewPath("/")
	if err != nil {
		t.Errorf("expected no error for valid path but got %v", err)
	}

	if !p.IsRoot() {
		t.Errorf("expected Path to be Root")
	}
}

func TestPath_Parent(t *testing.T) {
	p, err := NewPath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("expected no error for valid path but got %v", err)
	}

	p, err = p.Parent()
	if err != nil {
		t.Errorf("expected no error for Path with parent but got %v", err)
	}

	expectedPath := "/foo/Bar"
	if p.String() != expectedPath {
		t.Errorf("expected path %v but got %v", expectedPath, string(p))
	}

	p, err = p.Parent()
	if err != nil {
		t.Errorf("expected no error for Path with parent but got %v", err)
	}

	expectedPath = "/foo"
	if p.String() != expectedPath {
		t.Errorf("expected path %v but got %v", expectedPath, string(p))
	}

	p, err = p.Parent()
	if err != nil {
		t.Errorf("expected no error for Path with parent but got %v", err)
	}

	expectedPath = "/"
	if p.String() != expectedPath {
		t.Errorf("expected path %v but got %v", expectedPath, string(p))
	}

	_, err = p.Parent()
	if err == nil {
		t.Errorf("expected error for Path with no parent but got nil")
	}
}

func TestPath_Leaf(t *testing.T) {
	p := RootPath()
	_, err := p.Leaf()
	if err == nil {
		t.Errorf("expected error for root leaf but got none")
	}

	p = RootPath().Child("a")
	leaf, err := p.Leaf()
	if err != nil {
		t.Errorf("expected no error for short leaf but got %v", err)
	}
	if leaf != "a" {
		t.Errorf("expected leaf to be a but got %v", leaf)
	}

	p = RootPath().Child("c").Child("b").Child("a")
	leaf, err = p.Leaf()
	if err != nil {
		t.Errorf("expected no error for longer leaf but got %v", err)
	}
	if leaf != "a" {
		t.Errorf("expected leaf to be a but got %v", leaf)
	}
}

func TestPath_Prefix(t *testing.T) {
	p := RootPath().Child("a").Child("b").Child("c")
	_, err := p.Prefix(4)
	if err == nil {
		t.Errorf("expected error for long prefix but got none")
	}

	p2, err := p.Prefix(3)
	if err != nil {
		t.Errorf("expected no error for same length prefix but got one")
	}

	if p2 != p {
		t.Errorf("expected paths to be the same for same length prefix but they were different (expected %v, got %v)", p.String(), p2.String())
	}

	p2, err = p.Prefix(2)
	if err != nil {
		t.Errorf("expected no error for smaller prefix but got one")
	}

	if p2 != RootPath().Child("a").Child("b") {
		t.Errorf("unexpected prefix (expected %v, got %v)", RootPath().Child("a").Child("b").String(), p2.String())
	}

	p2, err = p.Prefix(1)
	if err != nil {
		t.Errorf("expected no error for smaller prefix but got one")
	}

	if p2 != RootPath().Child("a") {
		t.Errorf("unexpected prefix (expected %v, got %v)", RootPath().Child("a").String(), p2.String())
	}

	p2, err = p.Prefix(0)
	if err != nil {
		t.Errorf("expected no error for 0 prefix but got one")
	}

	if p2 != RootPath() {
		t.Errorf("unexpected prefix (expected %v, got %v)", p.String(), p2.String())
	}

	_, err = p.Prefix(-1)
	if err == nil {
		t.Errorf("expected error for negative depth prefix but got none")
	}
}

func TestPath_HasPrefix(t *testing.T) {
	p := RootPath().Child("a").Child("b").Child("c")
	if !p.HasPrefix(p) {
		t.Errorf("expected path to have itself as a prefix")
	}

	if !p.HasPrefix(RootPath().Child("a").Child("b")) {
		t.Errorf("expected path to have its first two subpaths as a prefix")
	}

	if !p.HasPrefix(RootPath().Child("a")) {
		t.Errorf("expected path to have its first subpath as a prefix")
	}

	if !p.HasPrefix(RootPath()) {
		t.Errorf("expected path to have root as a prefix")
	}

	if p.HasPrefix(p.Child("d")) {
		t.Errorf("expected path to not have a longer path as a prefix")
	}

	parent, _ := p.Parent()
	if p.HasPrefix(parent.Child("d")) {
		t.Errorf("expected path not to have a different leaf as a prefix")
	}

	grandparent, _ := parent.Parent()
	if p.HasPrefix(grandparent.Child("d")) {
		t.Errorf("expected path to not have different parent as prefix")
	}
}

func TestPath_Shift(t *testing.T) {
	tests := []struct {
		d string
		p Path
		n int
		e bool
		r Path
		s []string
	}{
		{
			d: "shift 0",
			p: Path("/foo/bar/bazz/buzz"),
			n: 0,
			r: Path("/foo/bar/bazz/buzz"),
			s: []string{},
		},
		{
			d: "shift 1",
			p: Path("/foo/bar/bazz/buzz"),
			n: 1,
			r: Path("/bar/bazz/buzz"),
			s: []string{"foo"},
		},
		{
			d: "shift 2",
			p: Path("/foo/bar/bazz/buzz"),
			n: 2,
			r: Path("/bazz/buzz"),
			s: []string{"foo", "bar"},
		},
		{
			d: "shift to root",
			p: Path("/foo/bar/bazz/buzz"),
			n: 4,
			r: Path("/"),
			s: []string{"foo", "bar", "bazz", "buzz"},
		},
		{
			d: "shift past root",
			p: Path("/foo/bar/bazz/buzz"),
			n: 5,
			e: true,
		},
	}

	for _, test := range tests {
		r, s, err := test.p.Shift(test.n)
		if err != nil {
			if !test.e {
				t.Errorf("expected no error for %v but got %v", test.d, err)
			}
			continue
		}

		if !reflect.DeepEqual(r, test.r) {
			t.Errorf("expected %v but got %v for %v", test.r, r, test.d)
		}

		if !reflect.DeepEqual(s, test.s) {
			t.Errorf("expected shifted components %v but got %v for %v", test.s, s, test.d)
		}
	}
}

func TestPathJSON(t *testing.T) {
	p := RootPath().Child("a").Child("b")
	data, err := json.Marshal(&p)
	if err != nil {
		t.Fatalf("got error marshalling path: %v", err)
	}

	p2 := Path("")
	if err := json.Unmarshal(data, &p2); err != nil {
		t.Fatalf("got error unmarshalling path: %v", err)
	}

	if p != p2 {
		t.Errorf("marshalled path is not same as the original. expected %v, got %v", p.String(), p2.String())
	}
}

func TestNewPathSubcomponent(t *testing.T) {
	_, err := NewPathSubcomponent("/foo/Bar/BUZZ")
	if err == nil {
		t.Errorf("expected error for empty subcomponent but got nil")
	}

	_, err = NewPathSubcomponent("/foo/Bar/BUZZ:")
	if err == nil {
		t.Errorf("expected error for empty subcomponent but got nil")
	}

	_, err = NewPathSubcomponent("/foo/Bar/BUZZ:bazz")
	if err != nil {
		t.Errorf("expected no error for valid path but got %v", err)
	}
}

func TestPathSubcomponent_Parts(t *testing.T) {
	p := RootPath().Child("a").Child("b")
	n, err := NewPathSubcomponentFromParts(p, "bazz")
	if err != nil {
		t.Errorf("expected no error for valid path subcomponent but got %v", err)
	}

	path, component := n.Parts()
	if path != p {
		t.Errorf("expected path %v but got %v", p.String(), path.String())
	}

	expectedComponent := "bazz"
	if component != expectedComponent {
		t.Errorf("expected path %v but got %v", p.String(), path.String())
	}

	path = n.Path()
	if path != p {
		t.Errorf("expected path %v but got %v", p.String(), path.String())
	}

	if n.Subcomponent() != expectedComponent {
		t.Errorf("expected component %v but got %v", expectedComponent, n.Subcomponent())
	}
}

func TestPathSubcomponentJSON(t *testing.T) {
	p := RootPath().Child("a").Child("b")
	c, _ := NewPathSubcomponentFromParts(p, "foo")
	data, err := json.Marshal(&c)
	if err != nil {
		t.Fatalf("got error marshalling path: %v", err)
	}

	c2 := PathSubcomponent("")
	if err := json.Unmarshal(data, &c2); err != nil {
		t.Fatalf("got error unmarshalling path: %v", err)
	}

	if c != c2 {
		t.Errorf("marshalled path subcomponent is not same as the original. expected %v, got %v", c.String(), c2.String())
	}
}
