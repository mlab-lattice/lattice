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

func (cbc *ComponentBuildClient) Get() (*types.ComponentBuild, error) {
	build := &types.ComponentBuild{}
	err := cbc.restClient.Get(cbc.baseURL).JSON(&build)
	return build, err
}

func (cbc *ComponentBuildClient) Logs(follow bool) (io.ReadCloser, error) {
	log, err := cbc.restClient.Get(cbc.baseURL + fmt.Sprintf("%v?follow=%v", componentBuildLogSubpath, follow)).Body()
	return log, err
}
