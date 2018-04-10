package v1

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type (
	ServiceState string
)

const (
	ServiceStatePending     ServiceState = "pending"
	ServiceStateScalingDown ServiceState = "scaling down"
	ServiceStateScalingUp   ServiceState = "scaling up"
	ServiceStateUpdating    ServiceState = "updating"
	ServiceStateStable      ServiceState = "stable"
	ServiceStateFailed      ServiceState = "failed"
)

// FIXME: should we expose Service ID, or just Path?
type Service struct {
	Path             tree.NodePath      `json:"path"`
	State            ServiceState       `json:"state"`
	UpdatedInstances int32              `json:"updatedInstances"`
	StaleInstances   int32              `json:"staleInstances"`
	PublicPorts      ServicePublicPorts `json:"publicPorts"`
	FailureMessage   *string            `json:"failureMessage,omitempty"`
}

type (
	ServicePublicPorts map[int32]ServicePublicPort
	ServicePublicPort  struct {
		Address string `json:"address"`
	}
)
