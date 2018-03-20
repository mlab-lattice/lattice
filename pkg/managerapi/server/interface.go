package server

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

type Backend interface {
	// System
	CreateSystem(id types.SystemID, definitionURL string) (*types.System, error)
	ListSystems() ([]types.System, error)
	GetSystem(types.SystemID) (s *types.System, exists bool, err error)
	DeleteSystem(types.SystemID) error

	// SystemBuild
	BuildSystem(id types.SystemID, definitionRoot tree.Node, v types.SystemVersion) (types.BuildID, error)
	ListSystemBuilds(types.SystemID) ([]types.SystemBuild, error)
	GetSystemBuild(types.SystemID, types.BuildID) (b *types.SystemBuild, exists bool, err error)

	// ServiceBuild
	ListServiceBuilds(types.SystemID) ([]types.ServiceBuild, error)
	GetServiceBuild(types.SystemID, types.ServiceBuildID) (b *types.ServiceBuild, exists bool, err error)

	// ComponentBuild
	ListComponentBuilds(types.SystemID) ([]types.ComponentBuild, error)
	GetComponentBuild(types.SystemID, types.ComponentBuildID) (b *types.ComponentBuild, exists bool, err error)
	GetComponentBuildLogs(id types.SystemID, bid types.ComponentBuildID, follow bool) (rc io.ReadCloser, exists bool, err error)

	// Deploy
	RollOutSystemBuild(types.SystemID, types.BuildID) (types.DeployID, error)
	RollOutSystem(id types.SystemID, definitionRoot tree.Node, v types.SystemVersion) (types.DeployID, error)
	ListSystemRollouts(types.SystemID) ([]types.Deploy, error)
	GetSystemRollout(types.SystemID, types.DeployID) (r *types.Deploy, exists bool, err error)

	// SystemTeardown
	TearDownSystem(types.SystemID) (types.TeardownID, error)
	ListSystemTeardowns(types.SystemID) ([]types.SystemTeardown, error)
	GetSystemTeardown(types.SystemID, types.TeardownID) (t *types.SystemTeardown, exists bool, err error)

	// Service
	ListServices(types.SystemID) ([]types.Service, error)
	GetService(types.SystemID, tree.NodePath) (*types.Service, error)

	// Secret
	ListSecrets(types.SystemID) ([]types.Secret, error)
	GetSecret(system types.SystemID, path tree.NodePath, name string) (s *types.Secret, exists bool, err error)
	SetSecret(system types.SystemID, path tree.NodePath, name, value string) error
	UnsetSecret(system types.SystemID, path tree.NodePath, name string) error
}
