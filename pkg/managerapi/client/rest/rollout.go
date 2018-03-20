package rest

import (
	"fmt"

	"bytes"
	"encoding/json"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	rolloutSubpath = "/rollouts"
)

type RolloutClient struct {
	restClient rest.Client
	baseURL    string
}

func newRolloutClient(c rest.Client, baseURL string) *RolloutClient {
	return &RolloutClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, rolloutSubpath),
	}
}

func (c *RolloutClient) List() ([]types.Deploy, error) {
	var rollouts []types.Deploy
	err := c.restClient.Get(c.baseURL).JSON(&rollouts)
	return rollouts, err
}

func (c *RolloutClient) Get(id types.DeployID) (*types.Deploy, error) {
	Rollout := &types.Deploy{}
	err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&Rollout)
	return Rollout, err
}

type rollOutSystemRequest struct {
	Version *string        `json:"version,omitempty"`
	BuildID *types.BuildID `json:"buildId,omitempty"`
}

type rolloutResponse struct {
	RolloutID types.DeployID `json:"rolloutId"`
}

func (c *RolloutClient) CreateFromBuild(id types.BuildID) (types.DeployID, error) {
	request := rollOutSystemRequest{
		BuildID: &id,
	}
	return c.create(request)
}

func (c *RolloutClient) CreateFromVersion(version string) (types.DeployID, error) {
	request := rollOutSystemRequest{
		Version: &version,
	}
	return c.create(request)
}

func (c *RolloutClient) create(request rollOutSystemRequest) (types.DeployID, error) {
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	rolloutResponse := &rolloutResponse{}
	err = c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).JSON(&rolloutResponse)
	if err != nil {
		return "", err
	}

	return rolloutResponse.RolloutID, nil
}
