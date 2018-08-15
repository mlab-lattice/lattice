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
	// Service ID
	ID ServiceID `json:"id"`
	// Service Path
	Path tree.NodePath `json:"path"`
	// State
	State ServiceState `json:"state"`
	// Message
	Message *string `json:"message,omitempty"`
	// FailureInfo
	FailureInfo *ServiceFailureInfo `json:"failureInfo,omitempty"`
	// AvailableInstances
	AvailableInstances int32 `json:"availableInstances"`
	// UpdatedInstances
	UpdatedInstances int32 `json:"updatedInstances"`
	// StaleInstances
	StaleInstances int32 `json:"staleInstances"`
	// TerminatingInstances
	TerminatingInstances int32 `json:"terminatingInstances"`
	// Ports
	Ports map[int32]string `json:"ports"`
	// Instances
	Instances []string `json:"instances"`
}

type ServiceFailureInfo struct {
	// Time
	Time time.Time
	// Message
	Message string `json:"message"`
}
