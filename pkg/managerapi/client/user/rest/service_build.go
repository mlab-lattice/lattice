package rest

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

type ServiceBuildClient struct {
	restClient rest.Client
	baseURL    string
	id         types.ServiceBuildID
}

func newServiceBuildClient(c rest.Client, baseURL string, id types.ServiceBuildID) *ServiceBuildClient {
	return &ServiceBuildClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v/%v", baseURL, serviceBuildSubpath, id),
		id:         id,
	}
}

func (sbc *ServiceBuildClient) Get() (*types.ServiceBuild, error) {
	build := &types.ServiceBuild{}
	err := sbc.restClient.Get(sbc.baseURL).JSON(&build)
	return build, err
}
