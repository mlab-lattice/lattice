package v1

import (
	"time"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type (
	ServiceID    string
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
	ID   ServiceID     `json:"id"`
	Path tree.NodePath `json:"path"`

	State       ServiceState        `json:"state"`
	Message     *string             `json:"message,omitempty"`
	FailureInfo *ServiceFailureInfo `json:"failureInfo,omitempty"`

	AvailableInstances   int32 `json:"availableInstances"`
	UpdatedInstances     int32 `json:"updatedInstances"`
	StaleInstances       int32 `json:"staleInstances"`
	TerminatingInstances int32 `json:"terminatingInstances"`

	Ports map[int32]string `json:"ports"`

	Instances []string `json:"instances"`
}

type ServiceFailureInfo struct {
	Time    time.Time
	Message string `json:"message"`
}
