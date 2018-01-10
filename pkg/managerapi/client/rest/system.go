package rest

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	systemSubpath = "/systems"
)

type SystemClient struct {
	restClient rest.Client
	baseURL    string
}

func newSystemClient(c rest.Client, baseURL string) client.SystemClient {
	return &SystemClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, systemSubpath),
	}
}

func (c *SystemClient) List() ([]types.System, error) {
	var systems []types.System
	err := c.restClient.Get(c.baseURL).JSON(&systems)
	return systems, err
}

func (c *SystemClient) Get(id types.SystemID) (*types.System, error) {
	system := &types.System{}
	err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&system)
	return system, err
}

func (c *SystemClient) SystemBuilds(id types.SystemID) client.SystemBuildClient {
	return newSystemBuildClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}

func (c *SystemClient) ServiceBuilds(id types.SystemID) client.ServiceBuildClient {
	return newServiceBuildClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}

func (c *SystemClient) ComponentBuilds(id types.SystemID) client.ComponentBuildClient {
	return newComponentBuildClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}

func (c *SystemClient) Rollouts(id types.SystemID) client.RolloutClient {
	return newRolloutClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}

func (c *SystemClient) Teardowns(id types.SystemID) client.TeardownClient {
	return newTeardownClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}

func (c *SystemClient) Services(id types.SystemID) client.ServiceClient {
	return newServiceClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}
