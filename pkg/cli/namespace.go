package cli

import (
	"github.com/mlab-lattice/system/pkg/cli/namespace"
	"github.com/mlab-lattice/system/pkg/cli/resources"
	"github.com/mlab-lattice/system/pkg/managerapi/client/user"
)

type NamespaceClient interface {
	ComponentBuilds() namespace.ComponentBuildClient
	ServiceBuilds() namespace.ServiceBuildClient
	SystemBuilds() namespace.SystemBuildClient
}

func NewNamespaceClient(ns user.NamespaceClient, asJSON bool) NamespaceClient {
	return &namespace.NamespaceClient{
		BuildClient: resources.BuildClient{
			RestClient:    ns,
			DisplayAsJSON: asJSON,
		},
	}
}
