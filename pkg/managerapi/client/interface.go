package client

import (
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"
)

type Interface interface {
	Status() (bool, error)

	Systems() SystemClient
}

type SystemClient interface {
	Create(id types.SystemID, definitionURL string) (*types.System, error)
	List() ([]types.System, error)
	Get(types.SystemID) (*types.System, error)
	Delete(types.SystemID) error

	Builds(types.SystemID) BuildClient
	Deploys(types.SystemID) DeployClient
	Teardowns(types.SystemID) TeardownClient
	Services(types.SystemID) ServiceClient
	Secrets(types.SystemID) SecretClient
}

type BuildClient interface {
	Create(version string) (types.BuildID, error)
	List() ([]types.Build, error)
	Get(types.BuildID) (*types.Build, error)
}

type DeployClient interface {
	CreateFromBuild(types.BuildID) (types.DeployID, error)
	CreateFromVersion(string) (types.DeployID, error)
	List() ([]types.Deploy, error)
	Get(types.DeployID) (*types.Deploy, error)
}

type TeardownClient interface {
	Create() (types.TeardownID, error)
	List() ([]types.SystemTeardown, error)
	Get(types.TeardownID) (*types.SystemTeardown, error)
}

type ServiceClient interface {
	List() ([]types.Service, error)
	Get(types.ServiceID) (*types.Service, error)
}

type SecretClient interface {
	List() ([]types.Secret, error)
	Get(path tree.NodePath, name string) (*types.Secret, error)
	Set(path tree.NodePath, name, value string) error
	Unset(path tree.NodePath, name string) error
}
