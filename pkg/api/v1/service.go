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
	// State ["pending", "deleting", "scaling", "updating", "stable", "failed"]
	State ServiceState `json:"state"`
	// TBD
	Message *string `json:"message,omitempty"`
	// TBD
	FailureInfo *ServiceFailureInfo `json:"failureInfo,omitempty"`
	// TBD
	AvailableInstances int32 `json:"availableInstances"`
	// TBD
	UpdatedInstances int32 `json:"updatedInstances"`
	// TBD
	StaleInstances int32 `json:"staleInstances"`
	// TBD
	TerminatingInstances int32 `json:"terminatingInstances"`
	// TBD
	Ports map[int32]string `json:"ports"`
	// TBD
	Instances []string `json:"instances"`
}

type ServiceFailureInfo struct {
	Time    time.Time
	Message string `json:"message"`
}
