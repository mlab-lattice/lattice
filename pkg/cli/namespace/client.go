package namespace

import (
	"github.com/mlab-lattice/system/pkg/cli/namespace/componentbuilds"
	"github.com/mlab-lattice/system/pkg/cli/namespace/servicebuilds"
	"github.com/mlab-lattice/system/pkg/cli/namespace/systembuilds"
	"github.com/mlab-lattice/system/pkg/cli/resources"
	"github.com/mlab-lattice/system/pkg/types"
)

type ComponentBuildClient interface {
	Show(types.ComponentBuildID)
	List()
	GetLogs(types.ComponentBuildID, bool)
}

type ServiceBuildClient interface {
	Show(types.ServiceBuildID)
	List()
}

type SystemBuildClient interface {
	Show(types.SystemBuildID)
	List()
}

type NamespaceClient struct {
	resources.BuildClient
}

func (ns *NamespaceClient) getBuildClient() resources.BuildClient {
	return resources.BuildClient{
		RestClient:    ns.RestClient,
		DisplayAsJSON: ns.DisplayAsJSON,
	}
}

func (ns *NamespaceClient) ComponentBuilds() ComponentBuildClient {
	return componentbuilds.ComponentBuildClient{
		BuildClient: ns.getBuildClient(),
	}
}

func (ns *NamespaceClient) ServiceBuilds() ServiceBuildClient {
	return servicebuilds.ServiceBuildClient{
		BuildClient: ns.getBuildClient(),
	}
}

func (ns *NamespaceClient) SystemBuilds() SystemBuildClient {
	return systembuilds.SystemBuildClient{
		BuildClient: ns.getBuildClient(),
	}
}
