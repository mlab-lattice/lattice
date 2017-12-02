package tree

import (
	"testing"
)

func TestNewNodePath(t *testing.T) {
	_, err := NewNodePath("")
	if err == nil {
		t.Errorf("Expected error for empty string path but got nil")
	}

	_, err = NewNodePath("foo/bar")
	if err == nil {
		t.Errorf("Expected error for path not beginning with '/' but got nil")
	}

	_, err = NewNodePath("/foo//bar")
	if err == nil {
		t.Errorf("Expected error for path with emtpy subpath in middle of path but got nil")
	}

	_, err = NewNodePath("/foo/bar/")
	if err == nil {
		t.Errorf("Expected error for path with emtpy subpath at the end of the path but got nil")
	}

	_, err = NewNodePath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}
}

func TestNodePathFromDomain(t *testing.T) {
	p2, err := NodePathFromDomain("BUZZ.Bar.foo")
	if err != nil {
		t.Errorf("Expected no error for valid NodePathFromDomain but got %v", err)
	}

	expectedPath := "/foo/Bar/BUZZ"
	if string(p2) != expectedPath {
		t.Errorf("Expected path %v but got %v", expectedPath, string(p2))
	}
}

func TestNodePath_ToDomain(t *testing.T) {
	p, err := NewNodePath("/foo/Bar/BUZZ")
	if err != nil {
		t.Errorf("Expected no error for valid path but got %v", err)
	}

	domain := p.ToDomain(true)
	expectedDomain := "buzz.bar.foo"
	if domain != expectedDomain {
		t.Errorf("Expected domain %v but got %v", expectedDomain, domain)
	}

	domain = p.ToDomain(false)
	expectedDomain = "BUZZ.Bar.foo"
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
	if string(p) != expectedPath {
		t.Errorf("Expected path %v but got %v", expectedPath, string(p))
	}

	p, err = p.Parent()
	if err != nil {
		t.Errorf("Expected no error for NodePath with parent but got %v", err)
	}

	expectedPath = "/foo"
	if string(p) != expectedPath {
		t.Errorf("Expected path %v but got %v", expectedPath, string(p))
	}

	_, err = p.Parent()
	if err == nil {
		t.Errorf("Expected error for NodePath with no parent but got nil")
	}
}
