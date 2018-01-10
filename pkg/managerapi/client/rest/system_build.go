package rest

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	systemBuildSubpath = "/system-builds"
)

type SystemBuildClient struct {
	restClient rest.Client
	baseURL    string
}

func newSystemBuildClient(c rest.Client, baseURL string) *SystemBuildClient {
	return &SystemBuildClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v/%v", baseURL, systemBuildSubpath),
	}
}

func (c *SystemBuildClient) List() ([]types.SystemBuild, error) {
	var builds []types.SystemBuild
	err := c.restClient.Get(c.baseURL).JSON(&builds)
	return builds, err
}

func (c *SystemBuildClient) Get(id types.SystemBuildID) (*types.SystemBuild, error) {
	build := &types.SystemBuild{}
	err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&build)
	return build, err
}
