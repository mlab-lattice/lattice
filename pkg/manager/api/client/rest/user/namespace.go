package user

import (
	"fmt"

	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/manager/api/client/rest/requester"
)

const (
	systemBuildSubpath    = "/system-builds"
	serviceBuildSubpath   = "/service-builds"
	componentBuildSubpath = "/component-builds"
)

type NamespaceClient struct {
	*Client
	namespace coretypes.LatticeNamespace
}

func newNamespaceClient(c *Client, namespace coretypes.LatticeNamespace) *NamespaceClient {
	return &NamespaceClient{
		Client:    c,
		namespace: namespace,
	}
}

func (nc *NamespaceClient) URL(endpoint string) string {
	return nc.Client.URL(fmt.Sprintf("%v/%v%v", namespaceEndpointPath, nc.namespace, endpoint))
}

func (nc *NamespaceClient) SystemBuilds() ([]coretypes.SystemBuild, error) {
	builds := []coretypes.SystemBuild{}
	err := requester.GetRequestBodyJSON(nc, systemBuildSubpath, &builds)
	return builds, err
}

func (nc *NamespaceClient) ServiceBuilds() ([]coretypes.ServiceBuild, error) {
	builds := []coretypes.ServiceBuild{}
	err := requester.GetRequestBodyJSON(nc, serviceBuildSubpath, &builds)
	return builds, err
}

func (nc *NamespaceClient) ComponentBuilds() ([]coretypes.ComponentBuild, error) {
	builds := []coretypes.ComponentBuild{}
	err := requester.GetRequestBodyJSON(nc, componentBuildSubpath, &builds)
	return builds, err
}

func (nc *NamespaceClient) SystemBuild(id coretypes.SystemBuildID) *SystemBuildClient {
	return newSystemBuildClient(nc, id)
}

func (nc *NamespaceClient) ServiceBuild(id coretypes.ServiceBuildID) *ServiceBuildClient {
	return newServiceBuildClient(nc, id)
}

func (nc *NamespaceClient) ComponentBuild(id coretypes.ComponentBuildID) *ComponentBuildClient {
	return newComponentBuildClient(nc, id)
}
