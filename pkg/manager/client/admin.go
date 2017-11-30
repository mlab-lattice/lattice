package client

import (
	"net/http"
)

const (
	adminEndpointPath          = "/admin"
	contentTypeApplicationJSON = "application/json"
)

type AdminClient struct {
	client        *http.Client
	managerAPIURL string
}

func NewClient(managerAPIURL string) *AdminClient {
	return &AdminClient{
		client:        http.DefaultClient,
		managerAPIURL: managerAPIURL,
	}
}

func (ac *AdminClient) httpClient() *http.Client {
	return ac.client
}

func (ac *AdminClient) url(endpoint string) string {
	return ac.managerAPIURL + adminEndpointPath + endpoint
}

func (ac *AdminClient) Master() *MasterClient {
	return newMasterClient(ac)
}
