package rest

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

type SystemBuildClient struct {
	restClient rest.Client
	baseURL    string
	id         types.SystemBuildID
}

func newSystemBuildClient(c rest.Client, baseURL string, id types.SystemBuildID) *SystemBuildClient {
	return &SystemBuildClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v/%v", baseURL, systemBuildSubpath, id),
		id:         id,
	}
}

func (c *SystemBuildClient) Get() (*types.SystemBuild, error) {
	build := &types.SystemBuild{}
	err := c.restClient.Get(c.baseURL).JSON(&build)
	return build, err
}
