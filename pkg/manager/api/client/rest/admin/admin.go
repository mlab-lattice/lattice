package admin

import (
	"net/http"
)

const (
	adminEndpointPath = "/admin"
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

func (ac *Client) HTTPClient() *http.Client {
	return ac.httpClient
}

func (ac *Client) URL(endpoint string) string {
	return ac.managerAPIURL + adminEndpointPath + endpoint
}

func (ac *Client) Master() *MasterClient {
	return newMasterClient(ac)
}
