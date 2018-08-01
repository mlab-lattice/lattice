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

func TestNodePathFromDomain(t *testing.T) {
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

func TestNodePath_ToDomain(t *testing.T) {
	p, err := NewPath("/foo/Bar/BUZZ")
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
	p, err := NewPath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	depth := p.Depth()
	expectedDepth := 3
	if expectedDepth != depth {
		t.Errorf("Expected depth %v but got %v", expectedDepth, depth)
	}

	p, err = NewPath("/foo/Bar")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	depth = p.Depth()
	expectedDepth = 2
	if expectedDepth != depth {
		t.Errorf("Expected depth %v but got %v", expectedDepth, depth)
	}

	p, err = NewPath("/foo")
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
	p, err := NewPath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	if p.IsRoot() {
		t.Errorf("Expected Path to not be Root")
	}

	p, err = NewPath("/foo/Bar")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	if p.IsRoot() {
		t.Errorf("Expected Path to not be Root")
	}

	p, err = NewPath("/foo")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	if p.IsRoot() {
		t.Errorf("Expected Path to not be Root")
	}

	p, err = NewPath("/")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	if !p.IsRoot() {
		t.Errorf("Expected Path to be Root")
	}
}

func TestNodePath_Parent(t *testing.T) {
	p, err := NewPath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	p, err = p.Parent()
	if err != nil {
		t.Errorf("Expected no error for Path with parent but got %v", err)
	}

	expectedPath := "/foo/Bar"
	if p.String() != expectedPath {
		t.Errorf("Expected path %v but got %v", expectedPath, string(p))
	}

	p, err = p.Parent()
	if err != nil {
		t.Errorf("Expected no error for Path with parent but got %v", err)
	}

	expectedPath = "/foo"
	if p.String() != expectedPath {
		t.Errorf("Expected path %v but got %v", expectedPath, string(p))
	}

	p, err = p.Parent()
	if err != nil {
		t.Errorf("Expected no error for Path with parent but got %v", err)
	}

	expectedPath = "/"
	if p.String() != expectedPath {
		t.Errorf("Expected path %v but got %v", expectedPath, string(p))
	}

	_, err = p.Parent()
	if err == nil {
		t.Errorf("Expected error for Path with no parent but got nil")
	}
}

func TestNewNodePathSubcomponent(t *testing.T) {
	_, err := NewPathSubcomponentFromParts("/foo/Bar/BUZZ", "")
	if err == nil {
		t.Errorf("Expected error for empty subcomponent but got nil")
	}

	_, err = NewPathSubcomponentFromParts("/foo/Bar/BUZZ", "bazz")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}
}

func TestNodePathSubcomponentParts(t *testing.T) {
	n, err := NewPathSubcomponentFromParts("/foo/Bar/BUZZ", "bazz")
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
