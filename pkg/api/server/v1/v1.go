package v1

import (
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"
)

type Interface interface {
	// System
	CreateSystem(id v1.SystemID, definitionURL string) (*v1.System, error)
	ListSystems() ([]v1.System, error)
	GetSystem(v1.SystemID) (s *v1.System, exists bool, err error)
	DeleteSystem(v1.SystemID) error

	// Build
	Build(id v1.SystemID, definitionRoot tree.Node, v v1.SystemVersion) (v1.BuildID, error)
	ListBuilds(v1.SystemID) ([]v1.Build, error)
	GetBuild(v1.SystemID, v1.BuildID) (b *v1.Build, exists bool, err error)

	// Deploy
	DeployBuild(v1.SystemID, v1.BuildID) (v1.DeployID, error)
	DeployVersion(id v1.SystemID, definitionRoot tree.Node, v v1.SystemVersion) (v1.DeployID, error)
	ListDeploys(v1.SystemID) ([]v1.Deploy, error)
	GetDeploy(v1.SystemID, v1.DeployID) (r *v1.Deploy, exists bool, err error)

	// Teardown
	TearDown(v1.SystemID) (v1.TeardownID, error)
	ListTeardowns(v1.SystemID) ([]v1.SystemTeardown, error)
	GetTeardown(v1.SystemID, v1.TeardownID) (t *v1.SystemTeardown, exists bool, err error)

	// Service
	ListServices(v1.SystemID) ([]v1.Service, error)
	GetService(v1.SystemID, tree.NodePath) (*v1.Service, error)

	// Secret
	ListSecrets(v1.SystemID) ([]v1.Secret, error)
	GetSecret(system v1.SystemID, path tree.NodePath, name string) (s *v1.Secret, exists bool, err error)
	SetSecret(system v1.SystemID, path tree.NodePath, name, value string) error
	UnsetSecret(system v1.SystemID, path tree.NodePath, name string) error
}
