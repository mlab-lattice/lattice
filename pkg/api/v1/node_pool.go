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
