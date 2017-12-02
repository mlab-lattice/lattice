package user

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/manager/api/client/rest/requester"
	"github.com/mlab-lattice/system/pkg/types"
)

const (
	systemBuildSubpath    = "/system-builds"
	serviceBuildSubpath   = "/service-builds"
	componentBuildSubpath = "/component-builds"
)

type NamespaceClient struct {
	*Client
	namespace types.LatticeNamespace
}

func newNamespaceClient(c *Client, namespace types.LatticeNamespace) *NamespaceClient {
	return &NamespaceClient{
		Client:    c,
		namespace: namespace,
	}
}

func (nc *NamespaceClient) URL(endpoint string) string {
	return nc.Client.URL(fmt.Sprintf("%v/%v%v", namespaceEndpointPath, nc.namespace, endpoint))
}

func (nc *NamespaceClient) SystemBuilds() ([]types.SystemBuild, error) {
	builds := []types.SystemBuild{}
	err := requester.GetRequestBodyJSON(nc, systemBuildSubpath, &builds)
	return builds, err
}

func (nc *NamespaceClient) ServiceBuilds() ([]types.ServiceBuild, error) {
	builds := []types.ServiceBuild{}
	err := requester.GetRequestBodyJSON(nc, serviceBuildSubpath, &builds)
	return builds, err
}

func (nc *NamespaceClient) ComponentBuilds() ([]types.ComponentBuild, error) {
	builds := []types.ComponentBuild{}
	err := requester.GetRequestBodyJSON(nc, componentBuildSubpath, &builds)
	return builds, err
}

func (nc *NamespaceClient) SystemBuild(id types.SystemBuildID) *SystemBuildClient {
	return newSystemBuildClient(nc, id)
}

func (nc *NamespaceClient) ServiceBuild(id types.ServiceBuildID) *ServiceBuildClient {
	return newServiceBuildClient(nc, id)
}

func (nc *NamespaceClient) ComponentBuild(id types.ComponentBuildID) *ComponentBuildClient {
	return newComponentBuildClient(nc, id)
}
