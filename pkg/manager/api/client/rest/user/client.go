package user

import (
	"net/http"

	coretypes "github.com/mlab-lattice/core/pkg/types"
)

const (
	namespaceEndpointPath = "/namespaces"
)

type Client struct {
	httpClient    *http.Client
	managerAPIURL string
}

func NewClient(managerAPIURL string) *Client {
	return &Client{
		httpClient:    http.DefaultClient,
		managerAPIURL: managerAPIURL,
	}
}

func (uc *Client) HTTPClient() *http.Client {
	return uc.httpClient
}

func (uc *Client) URL(endpoint string) string {
	return uc.managerAPIURL + endpoint
}

func (ac *Client) Namespace(namespace coretypes.LatticeNamespace) *NamespaceClient {
	return newNamespaceClient(ac, namespace)
}
