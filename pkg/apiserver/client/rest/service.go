package rest

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/apiserver/client"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	serviceSubpath = "/services"
)

type ServiceClient struct {
	restClient rest.Client
	baseURL    string
	systemID   types.SystemID
}

func newServiceClient(c rest.Client, baseURL string, systemID types.SystemID) *ServiceClient {
	return &ServiceClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, serviceSubpath),
		systemID:   systemID,
	}
}

func (c *ServiceClient) List() ([]types.Service, error) {
	var services []types.Service
	statusCode, err := c.restClient.Get(c.baseURL).JSON(&services)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, &client.InvalidSystemIDError{
			ID: c.systemID,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *ServiceClient) Get(id types.ServiceID) (*types.Service, error) {
	build := &types.Service{}
	statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&build)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return build, nil
	}

	if statusCode == http.StatusNotFound {
		// FIXME: need to differentiate between invalid service id and system id
		return nil, &client.InvalidServiceIDError{
			ID: id,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}
