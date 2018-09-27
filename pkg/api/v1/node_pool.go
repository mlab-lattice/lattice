package v1

import (
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

// swagger:model NodePool
type NodePool struct {
	ID   string                `json:"id"`
	Path tree.PathSubcomponent `json:"path"`

	InstanceType string `json:"instanceType"`
	NumInstances int32  `json:"numInstances"`

	Status NodePoolStatus `json:"status"`
}

type NodePoolStatus struct {
	State       NodePoolState        `json:"state"`
	FailureInfo *NodePoolFailureInfo `json:"failureInfo,omitempty"`

	// FIXME: how to deal with epochs?
	InstanceType string `json:"instanceType"`
	NumInstances int32  `json:"numInstances"`
}

type NodePoolFailureInfo struct {
	Time    time.Time
	Message string `json:"message"`
}
