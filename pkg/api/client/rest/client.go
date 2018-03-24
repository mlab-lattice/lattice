package rest

import (
	"fmt"
	"net/http"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

type Client struct {
	restClient rest.Client
	baseURL    string
}

func NewClient(apiServerURL string) *Client {
	return &Client{
		restClient: rest.NewClient(),
		baseURL:    apiServerURL,
	}
}

func (c *Client) Status() (bool, error) {
	resp, err := c.restClient.Get(fmt.Sprintf("%v/status", c.baseURL)).Do()
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

func (c *Client) Systems() clientv1.SystemClient {
	return newSystemClient(c.restClient, c.baseURL)
}
