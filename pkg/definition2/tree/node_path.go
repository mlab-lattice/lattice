package tree

import (
	"fmt"
	"strings"
)

const (
	// NodePathSeparator is the separator for the subpaths in a NodePath.
	NodePathSeparator = "/"

	// NodePathSubcomponentSeparator is the separator between the NodePath and subcomponent
	// of a NodePathSubcomponent.
	NodePathSubcomponentSeparator = ":"
)

// NodePath is a path from the root of a tree to a node in the tree.
// The components of the path are separated by slashes.
// For example: /a/b/c
type NodePath string

func RootNodePath() NodePath {
	return NodePath("/")
}

func ChildNodePath(path NodePath, child string) NodePath {
	return NodePath(fmt.Sprintf("%v/%v", path.String(), child))
}

// NewNodePath validates the string passed in and returns a NodePath.
func NewNodePath(p string) (NodePath, error) {
	if p == "" {
		return "", fmt.Errorf("cannot pass empty string as path")
	}

	parts := strings.Split(p, NodePathSeparator)
	if parts[0] != "" {
		return "", fmt.Errorf("path must start with '%v'", NodePathSeparator)
	}

	// allow for '/' (root), but other than that subpaths cannot be empty
	if len(parts) > 2 {
		for _, part := range parts[1:] {
			if part == "" {
				return "", fmt.Errorf("path cannot contain empty subpath")
			}
		}
	}

	return NodePath(p), nil
}

// NewNodePathFromDomain validates the domain passed in and converts it to a NodePath.
func NewNodePathFromDomain(d string) (NodePath, error) {
	s := strings.Split(string(d), ".")

	p := ""
	for i := len(s) - 1; i >= 0; i-- {
		p += NodePathSeparator + s[i]
	}

	return NewNodePath(p)
}

// ToDomain returns the domain version of the path.
// For example: /a/b/c returns c.b.a
// N.B.: panics if the NodePath is invalid (i.e. does not contain a slash)
func (np NodePath) ToDomain() string {
	subpaths := np.Subpaths()

	domain := ""
	first := true
	for i := len(subpaths) - 1; i >= 0; i-- {
		subpath := subpaths[i]
		if !first {
			domain += "."
		}

		domain += strings.ToLower(subpath)
		first = false
	}

	return domain
}

// Subpaths returns a slice of the paths making up the NodePath.
// For example: /a/b/c returns [a, b, c].
func (np NodePath) Subpaths() []string {
	subpaths := strings.Split(np.String(), NodePathSeparator)
	if len(subpaths) > 0 {
		return subpaths[1:]
	}

	return subpaths
}

// Depth returns the number of subpaths in the NodePath.
// For example: /a/b/c returns 3
func (np NodePath) Depth() int {
	subpaths := np.Subpaths()

	// Check for root ('/') case
	if len(subpaths) == 1 && subpaths[0] == "" {
		return 0
	}

	return len(subpaths)
}

// Parent returns the NodePath one level up from the NodePath if it is the root,
// and returns an error if it is the root.
// For example: /a/b/c returns /a/b
func (np NodePath) Parent() (NodePath, error) {
	if np.IsRoot() {
		return "", fmt.Errorf("NodePath %v does not have a parent", np.String())
	}

	if np.IsRoot() {
	}

	subpaths := np.Subpaths()
	parentSubpaths := subpaths[:len(subpaths)-1]
	return NewNodePath(fmt.Sprintf("%v%v", NodePathSeparator, strings.Join(parentSubpaths, NodePathSeparator)))
}

// IsRoot returns a bool indicating if the node is the root node (i.e. "/")
func (np NodePath) IsRoot() bool {
	return np.Depth() == 0
}

// String returns a string representation of the NodePath.
func (np NodePath) String() string {
	return string(np)
}

// NodePathSubcomponent contains a node path and the name of a subcomponent
// The subcomponent is separated from the node path with a colon.
// For example: /a/b/c:foo
type NodePathSubcomponent string

// NewNodePathSubcomponent returns a validated NodePathSubcomponent using the path and subcomponent.
func NewNodePathSubcomponent(path NodePath, subcomponent string) (NodePathSubcomponent, error) {
	if subcomponent == "" {
		return NodePathSubcomponent(""), fmt.Errorf("cannot pass empty string as subcomponent")
	}

	n := NodePathSubcomponent(
		fmt.Sprintf(
			"%v%v%v",
			path.String(),
			NodePathSubcomponentSeparator,
			subcomponent,
		),
	)
	return n, nil
}

// NodePath returns the NodePath of the NodePathSubcomponent.
// For example: /a/b/c:foo returns /a/b/c
func (n NodePathSubcomponent) NodePath() (NodePath, error) {
	path, _, err := n.Parts()
	return path, err
}

// NodePath returns the subcomponent of the NodePathSubcomponent.
// For example: /a/b/c:foo returns foo
func (n NodePathSubcomponent) Subcomponent() (string, error) {
	_, subcomponent, err := n.Parts()
	return subcomponent, err
}

// NodePath returns the NodePath and subcomponent of the NodePathSubcomponent.
// For example: /a/b/c:foo returns /a/b/c, foo
func (n NodePathSubcomponent) Parts() (NodePath, string, error) {
	parts := strings.Split(n.String(), NodePathSubcomponentSeparator)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("improperly formatted NodePathSubccomponent: %v", n)
	}

	path, err := NewNodePath(parts[0])
	if err != nil {
		return "", "", err
	}

	return path, parts[1], nil
}

// String returns a string representation of the NodePathSubcomponent
func (n NodePathSubcomponent) String() string {
	return string(n)
}
