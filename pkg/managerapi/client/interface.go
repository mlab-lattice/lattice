package client

import (
	"io"

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

	SystemBuilds(types.SystemID) SystemBuildClient
	ServiceBuilds(types.SystemID) ServiceBuildClient
	ComponentBuilds(types.SystemID) ComponentBuildClient
	Rollouts(types.SystemID) RolloutClient
	Teardowns(types.SystemID) TeardownClient
	Services(types.SystemID) ServiceClient
}

type SystemBuildClient interface {
	Create(version string) (types.SystemBuildID, error)
	List() ([]types.SystemBuild, error)
	Get(types.SystemBuildID) (*types.SystemBuild, error)
}

type ServiceBuildClient interface {
	List() ([]types.ServiceBuild, error)
	Get(types.ServiceBuildID) (*types.ServiceBuild, error)
}

type ComponentBuildClient interface {
	List() ([]types.ComponentBuild, error)
	Get(types.ComponentBuildID) (*types.ComponentBuild, error)
	Logs(id types.ComponentBuildID, follow bool) (io.ReadCloser, error)
}

type RolloutClient interface {
	CreateFromBuild(types.SystemBuildID) (types.SystemRolloutID, error)
	CreateFromVersion(string) (types.SystemRolloutID, error)
	List() ([]types.SystemRollout, error)
	Get(types.SystemRolloutID) (*types.SystemRollout, error)
}

type TeardownClient interface {
	Create() (types.SystemTeardownID, error)
	List() ([]types.SystemTeardown, error)
	Get(types.SystemTeardownID) (*types.SystemTeardown, error)
}

type ServiceClient interface {
	List() ([]types.Service, error)
	Get(types.ServiceID) (*types.Service, error)
}
