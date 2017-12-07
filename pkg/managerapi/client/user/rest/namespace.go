package rest

import (
	"fmt"

	clientinterface "github.com/mlab-lattice/system/pkg/managerapi/client/user"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	namespaceEndpointPath = "/namespaces"
	systemBuildSubpath    = "/system-builds"
	serviceBuildSubpath   = "/service-builds"
	componentBuildSubpath = "/component-builds"
)

type NamespaceClient struct {
	restClient rest.Client
	baseURL    string
	namespace  types.LatticeNamespace
}

func newNamespaceClient(c rest.Client, baseURL string, namespace types.LatticeNamespace) clientinterface.NamespaceClient {
	return &NamespaceClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v/%v", baseURL, namespaceEndpointPath, namespace),
		namespace:  namespace,
	}
}

func (nc *NamespaceClient) SystemBuilds() ([]types.SystemBuild, error) {
	builds := []types.SystemBuild{}
	err := nc.restClient.Get(nc.baseURL + systemBuildSubpath).JSON(&builds)
	return builds, err
}

func (nc *NamespaceClient) ServiceBuilds() ([]types.ServiceBuild, error) {
	builds := []types.ServiceBuild{}
	err := nc.restClient.Get(nc.baseURL + serviceBuildSubpath).JSON(&builds)
	return builds, err
}

func (nc *NamespaceClient) ComponentBuilds() ([]types.ComponentBuild, error) {
	builds := []types.ComponentBuild{}
	err := nc.restClient.Get(nc.baseURL + componentBuildSubpath).JSON(&builds)
	return builds, err
}

func (nc *NamespaceClient) SystemBuild(id types.SystemBuildID) clientinterface.SystemBuildClient {
	return newSystemBuildClient(nc.restClient, nc.baseURL, id)
}

func (nc *NamespaceClient) ServiceBuild(id types.ServiceBuildID) clientinterface.ServiceBuildClient {
	return newServiceBuildClient(nc.restClient, nc.baseURL, id)
}

func (nc *NamespaceClient) ComponentBuild(id types.ComponentBuildID) clientinterface.ComponentBuildClient {
	return newComponentBuildClient(nc.restClient, nc.baseURL, id)
}
