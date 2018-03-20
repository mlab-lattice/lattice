package rest

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	teardownSubpath = "/teardowns"
)

type TeardownClient struct {
	restClient rest.Client
	baseURL    string
}

func newTeardownClient(c rest.Client, baseURL string) *TeardownClient {
	return &TeardownClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, teardownSubpath),
	}
}

func (c *TeardownClient) List() ([]types.SystemTeardown, error) {
	var teardowns []types.SystemTeardown
	err := c.restClient.Get(c.baseURL).JSON(&teardowns)
	return teardowns, err
}

func (c *TeardownClient) Get(id types.TeardownID) (*types.SystemTeardown, error) {
	teardown := &types.SystemTeardown{}
	err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&teardown)
	return teardown, err
}

type teardownResponse struct {
	TeardownID types.TeardownID `json:"teardownId"`
}

func (c *TeardownClient) Create() (types.TeardownID, error) {
	teardownResponse := &teardownResponse{}
	err := c.restClient.PostJSON(c.baseURL, nil).JSON(&teardownResponse)
	if err != nil {
		return "", err
	}

	return teardownResponse.TeardownID, nil
}
