package user

import (
	"fmt"

	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/manager/client/requester"
)

const (
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

func (nc *NamespaceClient) ComponentBuilds() ([]coretypes.ComponentBuild, error) {
	builds := []coretypes.ComponentBuild{}
	err := requester.GetRequestBodyJSON(nc, componentBuildSubpath, &builds)
	return builds, err
}

func (nc *NamespaceClient) ComponentBuild(id coretypes.ComponentBuildID) *ComponentBuildClient {
	return newComponentBuildClient(nc, id)
}
