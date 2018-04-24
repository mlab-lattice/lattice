package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type (
	ServiceState string
)

const (
	ServiceStatePending  ServiceState = "pending"
	ServiceStateScaling  ServiceState = "scaling"
	ServiceStateUpdating ServiceState = "updating"
	ServiceStateStable   ServiceState = "stable"
	ServiceStateFailed   ServiceState = "failed"
)

type Service struct {
	Path tree.NodePath `json:"path"`

	State          ServiceState `json:"state"`
	FailureMessage *string      `json:"failureMessage,omitempty"`
	Reason         *string      `json:"reason,omitempty"`

	UpdatedInstances int32 `json:"updatedInstances"`
	StaleInstances   int32 `json:"staleInstances"`

	Ports map[int32]string
}
