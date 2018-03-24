package rest

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	serviceBuildSubpath = "/service-builds"
)

type ServiceBuildClient struct {
	restClient rest.Client
	baseURL    string
}

func newServiceBuildClient(c rest.Client, baseURL string) *ServiceBuildClient {
	return &ServiceBuildClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, serviceBuildSubpath),
	}
}

func (c *ServiceBuildClient) List() ([]types.ServiceBuild, error) {
	var builds []types.ServiceBuild
	_, err := c.restClient.Get(c.baseURL).JSON(&builds)
	return builds, err
}

func (c *ServiceBuildClient) Get(id types.ServiceBuildID) (*types.ServiceBuild, error) {
	build := &types.ServiceBuild{}
	_, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&build)
	return build, err
}
