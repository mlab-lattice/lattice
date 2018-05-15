package v1

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type NodePoolPath string

func (p NodePoolPath) String() string {
	return string(p)
}

func NewServiceNodePoolPath(path tree.NodePath) NodePoolPath {
	return NodePoolPath(path.String())
}

func NewSystemSharedNodePoolPath(path tree.NodePath, name string) NodePoolPath {
	return NodePoolPath(fmt.Sprintf("%v:%v", path.String(), name))
}

func ParseNodePoolPath(path NodePoolPath) (tree.NodePath, *string, error) {
	parts := strings.Split(path.String(), ":")
	if len(parts) > 2 || len(parts) == 0 {
		return "", nil, fmt.Errorf("invalid node pool path format")
	}

	p, err := tree.NewNodePath(parts[0])
	if err != nil {
		return "", nil, err
	}

	if len(parts) == 1 {
		return p, nil, nil
	}

	return p, &parts[1], nil
}

type NodePool struct {
	ID   string       `json:"id"`
	Path NodePoolPath `json:"path"`

	// FIXME: how to deal with epochs?
	InstanceType string `json:"instanceType"`
	NumInstances int32  `json:"numInstances"`

	State NodePoolState `json:"state"`
}

type NodePoolState string

const (
	NodePoolStatePending  = "pending"
	NodePoolStateScaling  = "scaling"
	NodePoolStateUpdating = "updating"
	NodePoolStateFailed   = "failed"
)
