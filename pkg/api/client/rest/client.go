package rest

import (
	"fmt"
	"net/http"

	v1restclient "github.com/mlab-lattice/lattice/pkg/api/client/rest/v1"
	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

type Client struct {
	restClient   rest.Client
	apiServerURL string
}

func NewClient(apiServerURL string) *Client {
	return &Client{
		restClient:   rest.NewClient(),
		apiServerURL: apiServerURL,
	}
}

func (c *Client) Health() (bool, error) {
	resp, err := c.restClient.Get(fmt.Sprintf("%v/health", c.apiServerURL)).Do()
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

func (c *Client) V1() v1client.Interface {
	return v1restclient.NewClient(c.restClient, c.apiServerURL)
}
