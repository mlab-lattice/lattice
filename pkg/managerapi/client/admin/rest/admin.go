package rest

import (
	"github.com/mlab-lattice/system/pkg/managerapi/client/admin"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	adminEndpointPath = "/admin"
)

type Client struct {
	restClient rest.Client
	baseURL    string
}

func NewClient(managerAPIURL string) *Client {
	return &Client{
		restClient: rest.NewClient(),
		baseURL:    managerAPIURL + adminEndpointPath,
	}
}

func (ac *Client) Master() admin.MasterClient {
	return newMasterClient(ac.restClient, ac.baseURL)
}
