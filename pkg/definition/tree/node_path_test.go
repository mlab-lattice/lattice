package tree

import (
	"reflect"
	"testing"
)

func TestNewNodePath(t *testing.T) {
	tests := []struct {
		d string
		p string
		e bool
		r NodePath
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
			r: NodePath("/"),
		},
		{
			d: "valid path",
			p: "/foo/Bar/BUZZ",
			e: false,
			r: NodePath("/foo/Bar/BUZZ"),
		},
	}

	for _, test := range tests {
		p, err := NewNodePath(test.p)
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

func TestNodePathFromDomain(t *testing.T) {
	tests := []struct {
		d string
		p string
		e bool
		r NodePath
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
			r: NodePath("/foo/Bar/BUZZ"),
		},
	}

	for _, test := range tests {
		p, err := NewNodePathFromDomain(test.p)
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

func TestNodePath_ToDomain(t *testing.T) {
	p, err := NewNodePath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	domain := p.ToDomain()
	expectedDomain := "buzz.bar.foo"
	if domain != expectedDomain {
		t.Errorf("Expected domain %v but got %v", expectedDomain, domain)
	}
}

func TestNodePath_Depth(t *testing.T) {
	p, err := NewNodePath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	depth := p.Depth()
	expectedDepth := 3
	if expectedDepth != depth {
		t.Errorf("Expected depth %v but got %v", expectedDepth, depth)
	}

	p, err = NewNodePath("/foo/Bar")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	depth = p.Depth()
	expectedDepth = 2
	if expectedDepth != depth {
		t.Errorf("Expected depth %v but got %v", expectedDepth, depth)
	}

	p, err = NewNodePath("/foo")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	depth = p.Depth()
	expectedDepth = 1
	if expectedDepth != depth {
		t.Errorf("Expected depth %v but got %v", expectedDepth, depth)
	}
}

func TestNodePath_IsRoot(t *testing.T) {
	p, err := NewNodePath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	if p.IsRoot() {
		t.Errorf("Expected NodePath to not be Root")
	}

	p, err = NewNodePath("/foo/Bar")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	if p.IsRoot() {
		t.Errorf("Expected NodePath to not be Root")
	}

	p, err = NewNodePath("/foo")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	if p.IsRoot() {
		t.Errorf("Expected NodePath to not be Root")
	}

	p, err = NewNodePath("/")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	if !p.IsRoot() {
		t.Errorf("Expected NodePath to be Root")
	}
}

func TestNodePath_Parent(t *testing.T) {
	p, err := NewNodePath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	p, err = p.Parent()
	if err != nil {
		t.Errorf("Expected no error for NodePath with parent but got %v", err)
	}

	expectedPath := "/foo/Bar"
	if p.String() != expectedPath {
		t.Errorf("Expected path %v but got %v", expectedPath, string(p))
	}

	p, err = p.Parent()
	if err != nil {
		t.Errorf("Expected no error for NodePath with parent but got %v", err)
	}

	expectedPath = "/foo"
	if p.String() != expectedPath {
		t.Errorf("Expected path %v but got %v", expectedPath, string(p))
	}

	p, err = p.Parent()
	if err != nil {
		t.Errorf("Expected no error for NodePath with parent but got %v", err)
	}

	expectedPath = "/"
	if p.String() != expectedPath {
		t.Errorf("Expected path %v but got %v", expectedPath, string(p))
	}

	_, err = p.Parent()
	if err == nil {
		t.Errorf("Expected error for NodePath with no parent but got nil")
	}
}

func TestNewNodePathSubcomponent(t *testing.T) {
	_, err := NewNodePathSubcomponentFromParts("/foo/Bar/BUZZ", "")
	if err == nil {
		t.Errorf("Expected error for empty subcomponent but got nil")
	}

	_, err = NewNodePathSubcomponentFromParts("/foo/Bar/BUZZ", "bazz")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}
}

func TestNodePathSubcomponentParts(t *testing.T) {
	n, err := NewNodePathSubcomponentFromParts("/foo/Bar/BUZZ", "bazz")
	if err != nil {
		t.Errorf("Expected no error for valid path subcomponent but got %v", err)
	}

	path, component := n.Parts()

	expectedPath := "/foo/Bar/BUZZ"
	if path.String() != expectedPath {
		t.Errorf("Expected path %v but got %v", expectedPath, path.String())
	}

	expectedComponent := "bazz"
	if component != expectedComponent {
		t.Errorf("Expected path %v but got %v", expectedPath, path.String())
	}
}
