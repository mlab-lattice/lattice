package tree

import (
	"encoding/json"
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

type InvalidNodePathError struct {
	Message string
}

func NewInvalidNodePathError(message string) *InvalidNodePathError {
	return &InvalidNodePathError{message}
}

func (e *InvalidNodePathError) Error() string {
	return e.Message
}

func RootNodePath() NodePath {
	return NodePath("/")
}

// NewNodePath validates the string passed in and returns a NodePath.
func NewNodePath(p string) (NodePath, error) {
	if p == "" {
		return "", NewInvalidNodePathError("cannot pass empty string as path")
	}

	parts := strings.Split(p, NodePathSeparator)
	if parts[0] != "" {
		return "", NewInvalidNodePathError(fmt.Sprintf("path must start with '%v'", NodePathSeparator))
	}

	// allow for '/' (root), but other than that subpaths cannot be empty
	if len(parts) > 2 {
		for _, part := range parts[1:] {
			if part == "" {
				return "", NewInvalidNodePathError("path cannot contain empty subpath")
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

// Child returns the path of a child of the NodePath
// For example, child c of /a/b returns /a/b/c
func (p NodePath) Child(child string) NodePath {
	if p.IsRoot() {
		return NodePath(fmt.Sprintf("/%v", child))
	}

	return NodePath(fmt.Sprintf("%v/%v", p.String(), child))
}

// ToDomain returns the domain version of the path.
// For example: /a/b/c returns c.b.a
// N.B.: panics if the NodePath is invalid (i.e. does not contain a slash)
func (p NodePath) ToDomain() string {
	subpaths := p.Subpaths()

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
func (p NodePath) Subpaths() []string {
	subpaths := strings.Split(p.String(), NodePathSeparator)
	if len(subpaths) > 0 {
		return subpaths[1:]
	}

	return subpaths
}

// Depth returns the number of subpaths in the NodePath.
// For example: /a/b/c returns 3
func (p NodePath) Depth() int {
	subpaths := p.Subpaths()

	// Check for root ('/') case
	if len(subpaths) == 1 && subpaths[0] == "" {
		return 0
	}

	return len(subpaths)
}

// Parent returns the NodePath one level up from the NodePath if it is the root,
// and returns an error if it is the root.
// For example: /a/b/c returns /a/b
func (p NodePath) Parent() (NodePath, error) {
	if p.IsRoot() {
		return "", fmt.Errorf("NodePath %v does not have a parent", p.String())
	}

	if p.IsRoot() {
	}

	subpaths := p.Subpaths()
	parentSubpaths := subpaths[:len(subpaths)-1]
	return NewNodePath(fmt.Sprintf("%v%v", NodePathSeparator, strings.Join(parentSubpaths, NodePathSeparator)))
}

// IsRoot returns a bool indicating if the node is the root node (i.e. "/")
func (p NodePath) IsRoot() bool {
	return p.Depth() == 0
}

// String returns a string representation of the NodePath.
func (p NodePath) String() string {
	return string(p)
}

// UnmarshalJSON implements json.Unmarshaler
func (p *NodePath) UnmarshalJSON(data []byte) error {
	var val string
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}

	tmp, err := NewNodePath(val)
	if err != nil {
		return err
	}

	*p = tmp
	return nil
}

// NodePathSubcomponent contains a node path and the name of a subcomponent
// The subcomponent is separated from the node path with a colon.
// For example: /a/b/c:foo
type NodePathSubcomponent string

type InvalidNodePathSubcomponentError struct {
	Message string
}

func NewInvalidNodePathSubcomponentError(message string) *InvalidNodePathSubcomponentError {
	return &InvalidNodePathSubcomponentError{message}
}

func (e *InvalidNodePathSubcomponentError) Error() string {
	return e.Message
}

// NewNodePathSubcomponent returns a validated NodePathSubcomponent from the supplied value.
func NewNodePathSubcomponent(val string) (NodePathSubcomponent, error) {
	parts := strings.Split(val, NodePathSubcomponentSeparator)
	if len(parts) != 2 {
		return "", NewInvalidNodePathSubcomponentError(fmt.Sprintf("improperly formatted NodePathSubcomponent: %v", val))
	}

	path, err := NewNodePath(parts[0])
	if err != nil {
		return "", err
	}

	return NewNodePathSubcomponentFromParts(path, parts[1])
}

// NewNodePathSubcomponentFromParts returns a validated NodePathSubcomponent using the path and subcomponent.
func NewNodePathSubcomponentFromParts(path NodePath, subcomponent string) (NodePathSubcomponent, error) {
	if subcomponent == "" {
		return "", NewInvalidNodePathSubcomponentError("cannot pass empty string as subcomponent")
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
// N.B.: panics if the NodePathSubcomponent is improperly formed.
func (n NodePathSubcomponent) NodePath() NodePath {
	path, _ := n.Parts()
	return path
}

// NodePath returns the subcomponent of the NodePathSubcomponent.
// For example: /a/b/c:foo returns foo
// N.B.: panics if the NodePathSubcomponent is improperly formed.
func (n NodePathSubcomponent) Subcomponent() string {
	_, subcomponent := n.Parts()
	return subcomponent
}

// NodePath returns the NodePath and subcomponent of the NodePathSubcomponent.
// For example: /a/b/c:foo returns /a/b/c, foo
// N.B.: panics if the NodePathSubcomponent is improperly formed.
func (n NodePathSubcomponent) Parts() (NodePath, string) {
	parts := strings.Split(n.String(), NodePathSubcomponentSeparator)

	path, err := NewNodePath(parts[0])
	if err != nil {
		panic(err)
	}

	return path, parts[1]
}

// String returns a string representation of the NodePathSubcomponent
func (n NodePathSubcomponent) String() string {
	return string(n)
}

// UnmarshalJSON implements json.Unmarshaler
func (n *NodePathSubcomponent) UnmarshalJSON(data []byte) error {
	var val string
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}

	tmp, err := NewNodePathSubcomponent(val)
	if err != nil {
		return err
	}

	*n = tmp
	return nil
}
