package rest

import (
	"fmt"
	"io"

	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	componentBuildSubpath    = "/component-builds"
	componentBuildLogSubpath = "/logs"
)

type ComponentBuildClient struct {
	restClient rest.Client
	baseURL    string
}

func newComponentBuildClient(c rest.Client, baseURL string) *ComponentBuildClient {
	return &ComponentBuildClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, componentBuildSubpath),
	}
}

func (c *ComponentBuildClient) List() ([]types.ComponentBuild, error) {
	var builds []types.ComponentBuild
	err := c.restClient.Get(c.baseURL).JSON(&builds)
	return builds, err
}

func (c *ComponentBuildClient) Get(id types.ComponentBuildID) (*types.ComponentBuild, error) {
	build := &types.ComponentBuild{}
	err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&build)
	return build, err
}

func (c *ComponentBuildClient) Logs(id types.ComponentBuildID, follow bool) (io.ReadCloser, error) {
	url := fmt.Sprintf("%v/%v%v?follow=%v", c.baseURL, id, componentBuildLogSubpath, follow)
	log, err := c.restClient.Get(url).Body()
	return log, err
}
