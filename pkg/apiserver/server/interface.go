package server

import (
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

	// Build
	Build(id types.SystemID, definitionRoot tree.Node, v types.SystemVersion) (types.BuildID, error)
	ListBuilds(types.SystemID) ([]types.Build, error)
	GetBuild(types.SystemID, types.BuildID) (b *types.Build, exists bool, err error)

	// Deploy
	DeployBuild(types.SystemID, types.BuildID) (types.DeployID, error)
	DeployVersion(id types.SystemID, definitionRoot tree.Node, v types.SystemVersion) (types.DeployID, error)
	ListDeploys(types.SystemID) ([]types.Deploy, error)
	GetDeploy(types.SystemID, types.DeployID) (r *types.Deploy, exists bool, err error)

	// Teardown
	TearDown(types.SystemID) (types.TeardownID, error)
	ListTeardowns(types.SystemID) ([]types.SystemTeardown, error)
	GetTeardown(types.SystemID, types.TeardownID) (t *types.SystemTeardown, exists bool, err error)

	// Service
	ListServices(types.SystemID) ([]types.Service, error)
	GetService(types.SystemID, tree.NodePath) (*types.Service, error)

	// Secret
	ListSecrets(types.SystemID) ([]types.Secret, error)
	GetSecret(system types.SystemID, path tree.NodePath, name string) (s *types.Secret, exists bool, err error)
	SetSecret(system types.SystemID, path tree.NodePath, name, value string) error
	UnsetSecret(system types.SystemID, path tree.NodePath, name string) error
}
