package rest

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	serviceSubpath = "/services"
)

type ServiceClient struct {
	restClient rest.Client
	baseURL    string
}

func newServiceClient(c rest.Client, baseURL string) *ServiceClient {
	return &ServiceClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, serviceSubpath),
	}
}

func (c *ServiceClient) List() ([]types.Service, error) {
	var services []types.Service
	err := c.restClient.Get(c.baseURL).JSON(&services)
	return services, err
}

func (c *ServiceClient) Get(id types.ServiceID) (*types.Service, error) {
	build := &types.Service{}
	err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&build)
	return build, err
}
