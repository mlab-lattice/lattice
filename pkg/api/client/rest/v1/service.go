package v1

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	serviceSubpath = "/services"
)

type ServiceClient struct {
	restClient rest.Client
	baseURL    string
}

func newServiceClient(c rest.Client, baseURL string) *ServiceClient {
	return &ServiceClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, serviceSubpath),
	}
}

func (c *ServiceClient) List() ([]v1.Service, error) {
	body, statusCode, err := c.restClient.Get(c.baseURL).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var services []v1.Service
		err = rest.UnmarshalBodyJSON(body, &services)
		return services, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *ServiceClient) Get(id v1.ServiceID) (*v1.Service, error) {
	body, statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		service := &v1.Service{}
		err = rest.UnmarshalBodyJSON(body, &service)
		return service, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}
