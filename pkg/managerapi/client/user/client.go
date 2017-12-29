package user

import (
	"io"

	"github.com/mlab-lattice/system/pkg/types"
)

type Client interface {
	// TODO: add Namespaces() ([]types.LatticeNamespace, error)

	System(types.SystemID) SystemClient
}

type SystemClient interface {
	SystemBuilds() ([]types.SystemBuild, error)
	ServiceBuilds() ([]types.ServiceBuild, error)
	ComponentBuilds() ([]types.ComponentBuild, error)

	SystemBuild(id types.SystemBuildID) SystemBuildClient
	ServiceBuild(id types.ServiceBuildID) ServiceBuildClient
	ComponentBuild(id types.ComponentBuildID) ComponentBuildClient
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
