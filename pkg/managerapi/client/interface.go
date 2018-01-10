package client

import (
	"io"

	"github.com/mlab-lattice/system/pkg/types"
)

type Interface interface {
	Systems() ([]types.System, error)

	System(types.SystemID) SystemClient
}

type SystemClient interface {
	Get() (*types.System, error)

	SystemBuilds() ([]types.SystemBuild, error)
	ServiceBuilds() ([]types.ServiceBuild, error)
	ComponentBuilds() ([]types.ComponentBuild, error)
	Services() ([]types.Service, error)

	SystemBuild(id types.SystemBuildID) SystemBuildClient
	ServiceBuild(id types.ServiceBuildID) ServiceBuildClient
	ComponentBuild(id types.ComponentBuildID) ComponentBuildClient
	Service(id types.ServiceID) ServiceClient
}

type SystemBuildClient interface {
	Get() (*types.SystemBuild, error)
}

type ServiceBuildClient interface {
	Get() (*types.ServiceBuild, error)
}

type ComponentBuildClient interface {
	Get() (*types.ComponentBuild, error)
	Logs(follow bool) (io.ReadCloser, error)
}

type ServiceClient interface {
	Get() (*types.Service, error)
}
