package tree

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// PathSeparator is the separator for the subpaths in a Path.
	PathSeparator = "/"

	// PathSubcomponentSeparator is the separator between the Path and subcomponent
	// of a PathSubcomponent.
	PathSubcomponentSeparator = ":"
)

// Path is a path from the root of a tree to a node in the tree.
// The components of the path are separated by slashes.
// For example: /a/b/c
type Path string

type InvalidPathError struct {
	Message string
}

func NewInvalidNodePathError(message string) *InvalidPathError {
	return &InvalidPathError{message}
}

func (e *InvalidPathError) Error() string {
	return e.Message
}

func RootPath() Path {
	return Path("/")
}

// NewPath validates the string passed in and returns a Path.
func NewPath(p string) (Path, error) {
	if p == "" {
		return "", NewInvalidNodePathError("cannot pass empty string as path")
	}

	parts := strings.Split(p, PathSeparator)
	if parts[0] != "" {
		return "", NewInvalidNodePathError(fmt.Sprintf("path must start with '%v'", PathSeparator))
	}

	// allow for '/' (root), but other than that subpaths cannot be empty
	if len(parts) > 2 {
		for _, part := range parts[1:] {
			if part == "" {
				return "", NewInvalidNodePathError("path cannot contain empty subpath")
			}
		}
	}

	return Path(p), nil
}

// NewPathFromDomain validates the domain passed in and converts it to a Path.
func NewPathFromDomain(d string) (Path, error) {
	s := strings.Split(string(d), ".")

	p := ""
	for i := len(s) - 1; i >= 0; i-- {
		p += PathSeparator + s[i]
	}

	return NewPath(p)
}

// Child returns the path of a child of the Path
// For example, child c of /a/b returns /a/b/c
func (p Path) Child(child string) Path {
	if p.IsRoot() {
		return Path(fmt.Sprintf("/%v", child))
	}

	return Path(fmt.Sprintf("%v/%v", p.String(), child))
}

// ToDomain returns the domain version of the path.
// For example: /a/b/c returns c.b.a
// N.B.: panics if the Path is invalid (i.e. does not contain a slash)
func (p Path) ToDomain() string {
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

// Subpaths returns a slice of the paths making up the Path.
// For example: /a/b/c returns [a, b, c].
func (p Path) Subpaths() []string {
	subpaths := strings.Split(p.String(), PathSeparator)
	if len(subpaths) > 0 {
		return subpaths[1:]
	}

	return subpaths
}

// Depth returns the number of subpaths in the Path.
// For example: /a/b/c returns 3
func (p Path) Depth() int {
	subpaths := p.Subpaths()

	// Check for root ('/') case
	if len(subpaths) == 1 && subpaths[0] == "" {
		return 0
	}

	return len(subpaths)
}

// Parent returns the Path one level up from the Path if it is the root,
// and returns an error if it is the root.
// For example: /a/b/c returns /a/b
func (p Path) Parent() (Path, error) {
	if p.IsRoot() {
		return "", fmt.Errorf("path %v does not have a parent", p.String())
	}

	subpaths := p.Subpaths()
	parentSubpaths := subpaths[:len(subpaths)-1]
	return NewPath(fmt.Sprintf("%v%v", PathSeparator, strings.Join(parentSubpaths, PathSeparator)))
}

// IsRoot returns a bool indicating if the node is the root node (i.e. "/")
func (p Path) IsRoot() bool {
	return p.Depth() == 0
}

// Shift returns a new Path shifted n components left as well as the components that were shifted out.
// Returns an error if n > the path's length.
// For example: /a/b/c shifted 2 would return /c and [a, b]
func (p Path) Shift(n int) (Path, []string, error) {
	if n > p.Depth() {
		return "", nil, fmt.Errorf("cannot shift path with depth %v %v times", p.Depth(), n)
	}

	subpaths := p.Subpaths()
	shifted := strings.Join(subpaths[n:], PathSeparator)

	np, err := NewPath(fmt.Sprintf("%v%v", PathSeparator, shifted))
	if err != nil {
		// this shouldn't happen if p is a valid Path
		return "", nil, err
	}

	return np, subpaths[:n], nil
}

// String returns a string representation of the Path.
func (p Path) String() string {
	return string(p)
}

// UnmarshalJSON implements json.Unmarshaler
func (p *Path) UnmarshalJSON(data []byte) error {
	var val string
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}

	tmp, err := NewPath(val)
	if err != nil {
		return err
	}

	*p = tmp
	return nil
}

// PathSubcomponent contains a node path and the name of a subcomponent
// The subcomponent is separated from the node path with a colon.
// For example: /a/b/c:foo
type PathSubcomponent string

type InvalidPathSubcomponentError struct {
	Message string
}

func NewInvalidPathSubcomponentError(message string) *InvalidPathSubcomponentError {
	return &InvalidPathSubcomponentError{message}
}

func (e *InvalidPathSubcomponentError) Error() string {
	return e.Message
}

// NewPathSubcomponent returns a validated PathSubcomponent from the supplied value.
func NewPathSubcomponent(val string) (PathSubcomponent, error) {
	parts := strings.Split(val, PathSubcomponentSeparator)
	if len(parts) != 2 {
		return "", NewInvalidPathSubcomponentError(fmt.Sprintf("improperly formatted PathSubcomponent: %v", val))
	}

	path, err := NewPath(parts[0])
	if err != nil {
		return "", err
	}

	return NewPathSubcomponentFromParts(path, parts[1])
}

// NewPathSubcomponentFromParts returns a validated PathSubcomponent using the path and subcomponent.
func NewPathSubcomponentFromParts(path Path, subcomponent string) (PathSubcomponent, error) {
	if subcomponent == "" {
		return "", NewInvalidPathSubcomponentError("cannot pass empty string as subcomponent")
	}

	n := PathSubcomponent(
		fmt.Sprintf(
			"%v%v%v",
			path.String(),
			PathSubcomponentSeparator,
			subcomponent,
		),
	)
	return n, nil
}

// Path returns the Path of the PathSubcomponent.
// For example: /a/b/c:foo returns /a/b/c
// N.B.: panics if the PathSubcomponent is improperly formed.
func (n PathSubcomponent) Path() Path {
	path, _ := n.Parts()
	return path
}

// Path returns the subcomponent of the PathSubcomponent.
// For example: /a/b/c:foo returns foo
// N.B.: panics if the PathSubcomponent is improperly formed.
func (n PathSubcomponent) Subcomponent() string {
	_, subcomponent := n.Parts()
	return subcomponent
}

// Path returns the Path and subcomponent of the PathSubcomponent.
// For example: /a/b/c:foo returns /a/b/c, foo
// N.B.: panics if the PathSubcomponent is improperly formed.
func (n PathSubcomponent) Parts() (Path, string) {
	parts := strings.Split(n.String(), PathSubcomponentSeparator)

	path, err := NewPath(parts[0])
	if err != nil {
		panic(err)
	}

	return path, parts[1]
}

// String returns a string representation of the PathSubcomponent
func (n PathSubcomponent) String() string {
	return string(n)
}

// UnmarshalJSON implements json.Unmarshaler
func (n *PathSubcomponent) UnmarshalJSON(data []byte) error {
	var val string
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}

	tmp, err := NewPathSubcomponent(val)
	if err != nil {
		return err
	}

	*n = tmp
	return nil
}
