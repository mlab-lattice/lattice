package rest

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	systemBuildSubpath    = "/system-builds"
	serviceBuildSubpath   = "/service-builds"
	componentBuildSubpath = "/component-builds"
	serviceSubpath        = "/services"
)

type SystemClient struct {
	restClient rest.Client
	baseURL    string
	systemID   types.SystemID
}

func newSystemClient(c rest.Client, baseURL string, systemID types.SystemID) client.SystemClient {
	return &SystemClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v/%v", baseURL, systemSubpath, systemID),
		systemID:   systemID,
	}
}

func (c *SystemClient) Get() (*types.System, error) {
	system := &types.System{}
	err := c.restClient.Get(c.baseURL).JSON(&system)
	return system, err
}

func (c *SystemClient) SystemBuilds() ([]types.SystemBuild, error) {
	var builds []types.SystemBuild
	err := c.restClient.Get(c.baseURL + systemBuildSubpath).JSON(&builds)
	return builds, err
}

func (c *SystemClient) ServiceBuilds() ([]types.ServiceBuild, error) {
	var builds []types.ServiceBuild
	err := c.restClient.Get(c.baseURL + serviceBuildSubpath).JSON(&builds)
	return builds, err
}

func (c *SystemClient) ComponentBuilds() ([]types.ComponentBuild, error) {
	var builds []types.ComponentBuild
	err := c.restClient.Get(c.baseURL + componentBuildSubpath).JSON(&builds)
	return builds, err
}

func (c *SystemClient) Services() ([]types.Service, error) {
	var services []types.Service
	err := c.restClient.Get(c.baseURL + componentBuildSubpath).JSON(&services)
	return services, err
}

func (c *SystemClient) SystemBuild(id types.SystemBuildID) client.SystemBuildClient {
	return newSystemBuildClient(c.restClient, c.baseURL, id)
}

func (c *SystemClient) ServiceBuild(id types.ServiceBuildID) client.ServiceBuildClient {
	return newServiceBuildClient(c.restClient, c.baseURL, id)
}

func (c *SystemClient) ComponentBuild(id types.ComponentBuildID) client.ComponentBuildClient {
	return newComponentBuildClient(c.restClient, c.baseURL, id)
}

func (c *SystemClient) Service(id types.ServiceID) client.ServiceClient {
	return newServiceClient(c.restClient, c.baseURL, id)
}
