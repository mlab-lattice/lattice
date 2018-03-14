package rest

import (
	"bytes"
	"encoding/json"
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
		baseURL:    fmt.Sprintf("%v%v", baseURL, systemBuildSubpath),
	}
}

type buildSystemRequest struct {
	Version string `json:"version,omitempty"`
}

type buildSystemResponse struct {
	BuildID types.SystemBuildID `json:"buildId"`
}

func (c *SystemBuildClient) Create(version string) (types.SystemBuildID, error) {
	request := &buildSystemRequest{
		Version: version,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	buildResponse := &buildSystemResponse{}
	err = c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).JSON(&buildSystemResponse{})
	if err != nil {
		return "", err
	}

	return buildResponse.BuildID, nil
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
