package backend

import (
	"io"

	systemtree "github.com/mlab-lattice/core/pkg/system/tree"
	coretypes "github.com/mlab-lattice/core/pkg/types"
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
	GetSystemUrl(coretypes.LatticeNamespace) (string, error)

	// Builds
	// System
	BuildSystem(ln coretypes.LatticeNamespace, definitionRoot systemtree.Node, v coretypes.SystemVersion) (coretypes.SystemBuildID, error)
	ListSystemBuilds(coretypes.LatticeNamespace) ([]coretypes.SystemBuild, error)
	GetSystemBuild(coretypes.LatticeNamespace, coretypes.SystemBuildID) (b *coretypes.SystemBuild, exists bool, err error)

	//Component
	ListComponentBuilds(coretypes.LatticeNamespace) ([]coretypes.ComponentBuild, error)
	GetComponentBuild(coretypes.LatticeNamespace, coretypes.ComponentBuildID) (b *coretypes.ComponentBuild, exists bool, err error)
	GetComponentBuildLogs(ln coretypes.LatticeNamespace, bid coretypes.ComponentBuildID, follow bool) (rc io.ReadCloser, exists bool, err error)

	// Rollouts
	RollOutSystemBuild(coretypes.LatticeNamespace, coretypes.SystemBuildID) (coretypes.SystemRolloutID, error)
	RollOutSystem(ln coretypes.LatticeNamespace, definitionRoot systemtree.Node, v coretypes.SystemVersion) (coretypes.SystemRolloutID, error)
	ListSystemRollouts(coretypes.LatticeNamespace) ([]coretypes.SystemRollout, error)
	GetSystemRollout(coretypes.LatticeNamespace, coretypes.SystemRolloutID) (r *coretypes.SystemRollout, exists bool, err error)

	// Teardowns
	TearDownSystem(coretypes.LatticeNamespace) (coretypes.SystemTeardownID, error)
	ListSystemTeardowns(coretypes.LatticeNamespace) ([]coretypes.SystemTeardown, error)
	GetSystemTeardown(coretypes.LatticeNamespace, coretypes.SystemTeardownID) (t *coretypes.SystemTeardown, exists bool, err error)

	// Services
	ListSystemServices(coretypes.LatticeNamespace) ([]coretypes.Service, error)
	GetSystemService(coretypes.LatticeNamespace, systemtree.NodePath) (*coretypes.Service, error)

	// Admin

	// Master Node Components
	GetMasterNodeComponents(nodeId string) ([]string, error)
	GetMasterNodeComponentLog(nodeId, componentName string, follow bool) (io.ReadCloser, error)
	RestartMasterNodeComponent(nodeId, componentName string) error
}
