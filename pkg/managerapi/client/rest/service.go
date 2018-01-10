package rest

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

type ServiceClient struct {
	restClient rest.Client
	baseURL    string
	id         types.ServiceID
}

func newServiceClient(c rest.Client, baseURL string, id types.ServiceID) *ServiceClient {
	return &ServiceClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v/%v", baseURL, serviceSubpath, id),
		id:         id,
	}
}

func (c *ServiceClient) Get() (*types.Service, error) {
	build := &types.Service{}
	err := c.restClient.Get(c.baseURL).JSON(&build)
	return build, err
}
