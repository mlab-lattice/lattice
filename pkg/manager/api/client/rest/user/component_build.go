package user

import (
	"fmt"
	"io"

	"github.com/mlab-lattice/system/pkg/manager/api/client/rest/requester"
	"github.com/mlab-lattice/system/pkg/types"
)

const (
	componentBuildLogSubpath = "/logs"
)

type ComponentBuildClient struct {
	*NamespaceClient
	id types.ComponentBuildID
}

func newComponentBuildClient(nc *NamespaceClient, id types.ComponentBuildID) *ComponentBuildClient {
	return &ComponentBuildClient{
		NamespaceClient: nc,
		id:              id,
	}
}

func (cbc *ComponentBuildClient) URL(endpoint string) string {
	return cbc.NamespaceClient.URL(fmt.Sprintf("%v/%v%v", componentBuildSubpath, cbc.id, endpoint))
}

func (cbc *ComponentBuildClient) Get() (*types.ComponentBuild, error) {
	build := &types.ComponentBuild{}
	err := requester.GetRequestBodyJSON(cbc, "", build)
	return build, err
}

func (cbc *ComponentBuildClient) Logs(follow bool) (io.ReadCloser, error) {
	log, err := requester.GetRequestBody(cbc, fmt.Sprintf("%v?follow=%v", componentBuildLogSubpath, follow))
	return log, err
}
