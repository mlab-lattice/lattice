package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type (
	ServiceState string
)

const (
	ServiceStatePending  ServiceState = "pending"
	ServiceStateDeleting ServiceState = "deleting"

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

	AvailableInstances   int32 `json:"availableInstances"`
	UpdatedInstances     int32 `json:"updatedInstances"`
	StaleInstances       int32 `json:"staleInstances"`
	TerminatingInstances int32 `json:"terminatingInstances"`

	Ports map[int32]string
}
