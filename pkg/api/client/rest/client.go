package rest

import (
	"fmt"
	"net/http"

	v1restclient "github.com/mlab-lattice/lattice/pkg/api/client/rest/v1"
	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

const (
	bearerTokenHeader = "Authorization"
)

type Client struct {
	restClient rest.Client
	url        string
}

func NewUnauthenticatedClient(url string) *Client {
	return &Client{
		restClient: rest.NewInsecureClient(nil),
		url:        url,
	}
}

func NewBearerTokenClient(url, bearerToken string) *Client {
	return &Client{
		restClient: rest.NewInsecureClient(
			map[string]string{bearerTokenHeader: fmt.Sprintf("Bearer %v", bearerToken)}),
		url: url,
	}
}

func (c *Client) Health() (bool, error) {
	resp, err := c.restClient.Get(fmt.Sprintf("%v/health", c.url)).Do()
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

func (c *Client) V1() v1client.Interface {
	return v1restclient.NewClient(c.restClient, c.url)
}
