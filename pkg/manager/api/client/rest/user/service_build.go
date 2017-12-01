package user

import (
	"fmt"

	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/manager/api/client/rest/requester"
)

type ServiceBuildClient struct {
	*NamespaceClient
	id coretypes.ServiceBuildID
}

func newServiceBuildClient(nc *NamespaceClient, id coretypes.ServiceBuildID) *ServiceBuildClient {
	return &ServiceBuildClient{
		NamespaceClient: nc,
		id:              id,
	}
}

func (sbc *ServiceBuildClient) URL(endpoint string) string {
	return sbc.NamespaceClient.URL(fmt.Sprintf("%v/%v%v", serviceBuildSubpath, sbc.id, endpoint))
}

func (sbc *ServiceBuildClient) Get() (*coretypes.ComponentBuild, error) {
	build := &coretypes.ComponentBuild{}
	err := requester.GetRequestBodyJSON(sbc, "", build)
	return build, err
}
