package backend

import (
	"io"

	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"
)

// UserError is an error that should be exposed to an end user.
type UserError struct {
	message string
}

// NewUserError constructs a new UserError
func NewUserError(message string) *UserError {
	return &UserError{
		message: message,
	}
}

// Error returns the UserError's message
func (e *UserError) Error() string {
	return e.message
}

type Interface interface {
	// Namespace

	// Utils
	GetSystemURL(types.LatticeNamespace) (string, error)

	// Builds
	// System
	BuildSystem(ln types.LatticeNamespace, definitionRoot tree.Node, v types.SystemVersion) (types.SystemBuildID, error)
	ListSystemBuilds(types.LatticeNamespace) ([]types.SystemBuild, error)
	GetSystemBuild(types.LatticeNamespace, types.SystemBuildID) (b *types.SystemBuild, exists bool, err error)

	// Service
	ListServiceBuilds(types.LatticeNamespace) ([]types.ServiceBuild, error)
	GetServiceBuild(types.LatticeNamespace, types.ServiceBuildID) (b *types.ServiceBuild, exists bool, err error)

	// Component
	ListComponentBuilds(types.LatticeNamespace) ([]types.ComponentBuild, error)
	GetComponentBuild(types.LatticeNamespace, types.ComponentBuildID) (b *types.ComponentBuild, exists bool, err error)
	GetComponentBuildLogs(ln types.LatticeNamespace, bid types.ComponentBuildID, follow bool) (rc io.ReadCloser, exists bool, err error)

	// Rollouts
	RollOutSystemBuild(types.LatticeNamespace, types.SystemBuildID) (types.SystemRolloutID, error)
	RollOutSystem(ln types.LatticeNamespace, definitionRoot tree.Node, v types.SystemVersion) (types.SystemRolloutID, error)
	ListSystemRollouts(types.LatticeNamespace) ([]types.SystemRollout, error)
	GetSystemRollout(types.LatticeNamespace, types.SystemRolloutID) (r *types.SystemRollout, exists bool, err error)

	// Teardowns
	TearDownSystem(types.LatticeNamespace) (types.SystemTeardownID, error)
	ListSystemTeardowns(types.LatticeNamespace) ([]types.SystemTeardown, error)
	GetSystemTeardown(types.LatticeNamespace, types.SystemTeardownID) (t *types.SystemTeardown, exists bool, err error)

	// Services
	ListSystemServices(types.LatticeNamespace) ([]types.Service, error)
	GetSystemService(types.LatticeNamespace, tree.NodePath) (*types.Service, error)

	// Admin

	// Master Node Components
	GetMasterNodeComponents(nodeID string) ([]string, error)
	GetMasterNodeComponentLog(nodeID, componentName string, follow bool) (rc io.ReadCloser, exists bool, err error)
	RestartMasterNodeComponent(nodeID, componentName string) (exists bool, err error)
}
