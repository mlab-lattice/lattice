package v1

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type NodePoolState string

const (
	NodePoolStatePending  = "pending"
	NodePoolStateDeleting = "deleting"

	NodePoolStateStable   = "stable"
	NodePoolStateScaling  = "scaling"
	NodePoolStateUpdating = "updating"
	NodePoolStateFailed   = "failed"
)

type NodePool struct {
	ID   string                `json:"id"`
	Path tree.PathSubcomponent `json:"path"`

	// FIXME: how to deal with epochs?
	InstanceType string `json:"instanceType"`
	NumInstances int32  `json:"numInstances"`

	State       NodePoolState        `json:"state"`
	FailureInfo *NodePoolFailureInfo `json:"failure_info"`
}

type NodePoolFailureInfo struct {
	Time    time.Time
	Message string `json:"message"`
}

type NodePoolPath struct {
	Path tree.Path `json:"path"`
	Name *string   `json:"name,omitempty"`
}

func (p NodePoolPath) String() string {
	if p.Name == nil {
		return p.Path.String()
	}

	return fmt.Sprintf("%v:%v", p.Path, *p.Name)
}

func (p NodePoolPath) MarshalJSON() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *NodePoolPath) UnmarshalJSON(data []byte) error {
	var path string
	err := json.Unmarshal(data, &path)
	if err != nil {
		return err
	}

	np, err := ParseNodePoolPath(path)
	if err != nil {
		return err
	}

	p.Path = np.Path
	p.Name = np.Name
	return nil
}

func NewServiceNodePoolPath(path tree.Path) NodePoolPath {
	return NodePoolPath{Path: path}
}

func NewSystemSharedNodePoolPath(path tree.Path, name string) NodePoolPath {
	return NodePoolPath{
		Path: path,
		Name: &name,
	}
}

func ParseNodePoolPath(path string) (NodePoolPath, error) {
	parts := strings.Split(path, ":")
	if len(parts) > 2 || len(parts) == 0 {
		return NodePoolPath{}, fmt.Errorf("invalid node pool path format")
	}

	p, err := tree.NewPath(parts[0])
	if err != nil {
		return NodePoolPath{}, err
	}

	if len(parts) == 1 {
		np := NodePoolPath{Path: p}
		return np, nil
	}

	np := NodePoolPath{
		Path: p,
		Name: &parts[1],
	}
	return np, nil
}
