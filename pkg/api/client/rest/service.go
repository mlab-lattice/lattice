package rest

import (
	"fmt"
	"net/http"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	serviceSubpath = "/services"
)

type ServiceClient struct {
	restClient rest.Client
	baseURL    string
	systemID   v1.SystemID
}

func newServiceClient(c rest.Client, baseURL string, systemID v1.SystemID) *ServiceClient {
	return &ServiceClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, serviceSubpath),
		systemID:   systemID,
	}
}

func (c *ServiceClient) List() ([]v1.Service, error) {
	var services []v1.Service
	statusCode, err := c.restClient.Get(c.baseURL).JSON(&services)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, &clientv1.InvalidSystemIDError{
			ID: c.systemID,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *ServiceClient) Get(id v1.ServiceID) (*v1.Service, error) {
	build := &v1.Service{}
	statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&build)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return build, nil
	}

	if statusCode == http.StatusNotFound {
		// FIXME: need to differentiate between invalid service id and system id
		return nil, &clientv1.InvalidServiceIDError{
			ID: id,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}
