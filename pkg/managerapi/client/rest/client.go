package rest

import (
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	systemSubpath = "/systems"
)

type Client struct {
	restClient rest.Client
	baseURL    string
}

func NewClient(managerAPIURL string) *Client {
	return &Client{
		restClient: rest.NewClient(),
		baseURL:    managerAPIURL,
	}
}

func (c *Client) Systems() ([]types.System, error) {
	var systems []types.System
	err := c.restClient.Get(c.baseURL + systemSubpath).JSON(&systems)
	return systems, err
}

func (c *Client) System(systemID types.SystemID) client.SystemClient {
	return newSystemClient(c.restClient, c.baseURL, systemID)
}
