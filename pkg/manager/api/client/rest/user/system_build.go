package user

import (
	"fmt"

	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/manager/api/client/rest/requester"
)

type SystemBuildClient struct {
	*NamespaceClient
	id coretypes.SystemBuildID
}

func newSystemBuildClient(nc *NamespaceClient, id coretypes.SystemBuildID) *SystemBuildClient {
	return &SystemBuildClient{
		NamespaceClient: nc,
		id:              id,
	}
}

func (sbc *SystemBuildClient) URL(endpoint string) string {
	return sbc.NamespaceClient.URL(fmt.Sprintf("%v/%v%v", systemBuildSubpath, sbc.id, endpoint))
}

func (sbc *SystemBuildClient) Get() (*coretypes.ComponentBuild, error) {
	build := &coretypes.ComponentBuild{}
	err := requester.GetRequestBodyJSON(sbc, "", build)
	return build, err
}
