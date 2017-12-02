package user

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/manager/api/client/rest/requester"
	"github.com/mlab-lattice/system/pkg/types"
)

type ServiceBuildClient struct {
	*NamespaceClient
	id types.ServiceBuildID
}

func newServiceBuildClient(nc *NamespaceClient, id types.ServiceBuildID) *ServiceBuildClient {
	return &ServiceBuildClient{
		NamespaceClient: nc,
		id:              id,
	}
}

func (sbc *ServiceBuildClient) URL(endpoint string) string {
	return sbc.NamespaceClient.URL(fmt.Sprintf("%v/%v%v", serviceBuildSubpath, sbc.id, endpoint))
}

func (sbc *ServiceBuildClient) Get() (*types.ComponentBuild, error) {
	build := &types.ComponentBuild{}
	err := requester.GetRequestBodyJSON(sbc, "", build)
	return build, err
}
