package rest

import (
	"fmt"

	clientinterface "github.com/mlab-lattice/system/pkg/managerapi/client/user"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	systemEndpointPath    = "/systems"
	systemBuildSubpath    = "/system-builds"
	serviceBuildSubpath   = "/service-builds"
	componentBuildSubpath = "/component-builds"
)

type SystemClient struct {
	restClient rest.Client
	baseURL    string
	systemID   types.SystemID
}

func newNamespaceClient(c rest.Client, baseURL string, systemID types.SystemID) clientinterface.SystemClient {
	return &SystemClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v/%v", baseURL, systemEndpointPath, systemID),
		systemID:   systemID,
	}
}

func (nc *SystemClient) SystemBuilds() ([]types.SystemBuild, error) {
	builds := []types.SystemBuild{}
	err := nc.restClient.Get(nc.baseURL + systemBuildSubpath).JSON(&builds)
	return builds, err
}

func (nc *SystemClient) ServiceBuilds() ([]types.ServiceBuild, error) {
	builds := []types.ServiceBuild{}
	err := nc.restClient.Get(nc.baseURL + serviceBuildSubpath).JSON(&builds)
	return builds, err
}

func (nc *SystemClient) ComponentBuilds() ([]types.ComponentBuild, error) {
	builds := []types.ComponentBuild{}
	err := nc.restClient.Get(nc.baseURL + componentBuildSubpath).JSON(&builds)
	return builds, err
}

func (nc *SystemClient) SystemBuild(id types.SystemBuildID) clientinterface.SystemBuildClient {
	return newSystemBuildClient(nc.restClient, nc.baseURL, id)
}

func (nc *SystemClient) ServiceBuild(id types.ServiceBuildID) clientinterface.ServiceBuildClient {
	return newServiceBuildClient(nc.restClient, nc.baseURL, id)
}

func (nc *SystemClient) ComponentBuild(id types.ComponentBuildID) clientinterface.ComponentBuildClient {
	return newComponentBuildClient(nc.restClient, nc.baseURL, id)
}
