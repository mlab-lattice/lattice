package tree

import (
	"fmt"
	"strings"
)

type NodePath string

func NewNodePath(p string) (NodePath, error) {
	if p == "" {
		return NodePath(""), fmt.Errorf("cannot pass empty string as path")
	}

	parts := strings.Split(p, "/")
	if parts[0] != "" {
		return NodePath(""), fmt.Errorf("path must start with '/'")
	}

	for _, part := range parts[1:] {
		if part == "" {
			return NodePath(""), fmt.Errorf("path cannot contain empty subpath")
		}
	}

	return NodePath(p), nil
}

func NodePathFromDomain(d string) (NodePath, error) {
	s := strings.Split(string(d), ".")

	p := ""
	for i := len(s) - 1; i >= 0; i-- {
		p += "/" + s[i]
	}

	return NewNodePath(p)
}

func (np NodePath) ToDomain() string {
	// FIXME: this will panic if it's an invalid path
	s := strings.Split(string(np), "/")[1:]

	domain := ""
	first := true
	for i := len(s) - 1; i >= 0; i-- {
		subpath := s[i]
		if !first {
			domain += "."
		}

		domain += strings.ToLower(subpath)
		first = false
	}

	return domain
}

func (np NodePath) Depth() int {
	return len(strings.Split(string(np), "/")) - 1
}

func (np NodePath) Parent() (NodePath, error) {
	if np.Depth() == 1 {
		return NodePath(""), fmt.Errorf("NodePath %v does not have a parent", np)
	}

	parts := strings.Split(string(np), "/")
	parentParts := parts[:len(parts)-1]
	return NodePath(strings.Join(parentParts, "/")), nil
}

func (np NodePath) IsRoot() bool {
	return np.Depth() == 1
}

func (np NodePath) String() string {
	return string(np)
}
