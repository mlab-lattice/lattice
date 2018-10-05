package v1

import (
	"github.com/mlab-lattice/lattice/pkg/util/time"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type (
	NodePoolID    string
	NodePoolState string
)

const (
	NodePoolStatePending  = "pending"
	NodePoolStateDeleting = "deleting"

	NodePoolStateStable   = "stable"
	NodePoolStateScaling  = "scaling"
	NodePoolStateUpdating = "updating"
	NodePoolStateFailed   = "failed"
)

type NodePool struct {
	ID   NodePoolID            `json:"id"`
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
