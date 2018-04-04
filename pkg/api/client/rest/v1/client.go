package v1

import (
	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

type Client struct {
	restClient   rest.Client
	apiServerURL string
}

func NewClient(client rest.Client, apiServerURL string) *Client {
	return &Client{
		restClient:   client,
		apiServerURL: apiServerURL,
	}
}

func (c *Client) Systems() v1client.SystemClient {
	return newSystemClient(c.restClient, c.apiServerURL)
}
