package types

import (
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type ServiceID string
type ServiceState string

const (
	ServiceStatePending     ServiceState = "pending"
	ServiceStateScalingDown ServiceState = "scaling down"
	ServiceStateScalingUp   ServiceState = "scaling up"
	ServiceStateUpdating    ServiceState = "updating"
	ServiceStateStable      ServiceState = "stable"
	ServiceStateFailed      ServiceState = "failed"
)

type Service struct {
	ID               ServiceID          `json:"id"`
	Path             tree.NodePath      `json:"path"`
	State            ServiceState       `json:"state"`
	UpdatedInstances int32              `json:"updatedInstances"`
	StaleInstances   int32              `json:"staleInstances"`
	PublicPorts      ServicePublicPorts `json:"publicPorts"`
	FailureMessage   *string            `json:"failureMessage"`
}

type ServicePublicPorts map[int32]ServicePublicPort

type ServicePublicPort struct {
	Address string `json:"address"`
}
