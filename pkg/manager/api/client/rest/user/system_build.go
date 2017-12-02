package user

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/manager/api/client/rest/requester"
	"github.com/mlab-lattice/system/pkg/types"
)

type SystemBuildClient struct {
	*NamespaceClient
	id types.SystemBuildID
}

func newSystemBuildClient(nc *NamespaceClient, id types.SystemBuildID) *SystemBuildClient {
	return &SystemBuildClient{
		NamespaceClient: nc,
		id:              id,
	}
}

func (sbc *SystemBuildClient) URL(endpoint string) string {
	return sbc.NamespaceClient.URL(fmt.Sprintf("%v/%v%v", systemBuildSubpath, sbc.id, endpoint))
}

func (sbc *SystemBuildClient) Get() (*types.ComponentBuild, error) {
	build := &types.ComponentBuild{}
	err := requester.GetRequestBodyJSON(sbc, "", build)
	return build, err
}
