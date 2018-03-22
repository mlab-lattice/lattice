package rest

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/apiserver/client"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

type Client struct {
	restClient rest.Client
	baseURL    string
}

func NewClient(aPIServerURL string) *Client {
	return &Client{
		restClient: rest.NewClient(),
		baseURL:    aPIServerURL,
	}
}

func (c *Client) Status() (bool, error) {
	resp, err := c.restClient.Get(fmt.Sprintf("%v/status", c.baseURL)).Do()
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

func (c *Client) Systems() client.SystemClient {
	return newSystemClient(c.restClient, c.baseURL)
}
