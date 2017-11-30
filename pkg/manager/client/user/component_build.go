package user

import (
	"fmt"
	"io"

	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/manager/client/requester"
)

const (
	componentBuildLogSubpath = "/logs"
)

type ComponentBuildClient struct {
	*NamespaceClient
	id coretypes.ComponentBuildID
}

func newComponentBuildClient(nc *NamespaceClient, id coretypes.ComponentBuildID) *ComponentBuildClient {
	return &ComponentBuildClient{
		NamespaceClient: nc,
		id:              id,
	}
}

func (cbc *ComponentBuildClient) URL(endpoint string) string {
	return cbc.NamespaceClient.URL(fmt.Sprintf("%v/%v%v", componentBuildSubpath, cbc.id, endpoint))
}

func (cbc *ComponentBuildClient) Get() (*coretypes.ComponentBuild, error) {
	build := &coretypes.ComponentBuild{}
	err := requester.GetRequestBodyJSON(cbc, "", build)
	return build, err
}

func (cbc *ComponentBuildClient) Logs(follow bool) (io.ReadCloser, error) {
	log, err := requester.GetRequestBody(cbc, fmt.Sprintf("%v?follow=%v", componentBuildLogSubpath, follow))
	return log, err
}
