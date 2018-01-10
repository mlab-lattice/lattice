package rest

import (
	"fmt"
	"io"

	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	componentBuildLogSubpath = "/logs"
)

type ComponentBuildClient struct {
	restClient rest.Client
	baseURL    string
	id         types.ComponentBuildID
}

func newComponentBuildClient(c rest.Client, baseURL string, id types.ComponentBuildID) *ComponentBuildClient {
	return &ComponentBuildClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v/%v", baseURL, componentBuildSubpath, id),
		id:         id,
	}
}

func (c *ComponentBuildClient) Get() (*types.ComponentBuild, error) {
	build := &types.ComponentBuild{}
	err := c.restClient.Get(c.baseURL).JSON(&build)
	return build, err
}

func (c *ComponentBuildClient) Logs(follow bool) (io.ReadCloser, error) {
	log, err := c.restClient.Get(c.baseURL + fmt.Sprintf("%v?follow=%v", componentBuildLogSubpath, follow)).Body()
	return log, err
}
