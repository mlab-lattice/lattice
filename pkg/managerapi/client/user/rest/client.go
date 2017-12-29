package rest

import (
	"github.com/mlab-lattice/system/pkg/managerapi/client/user"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
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

func (uc *Client) System(systemID types.SystemID) user.SystemClient {
	return newNamespaceClient(uc.restClient, uc.baseURL, systemID)
}
