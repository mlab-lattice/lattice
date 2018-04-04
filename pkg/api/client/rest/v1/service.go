package v1

import (
	"fmt"
	"net/http"
	urlutil "net/url"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

type ServiceClient struct {
	restClient   rest.Client
	apiServerURL string
	systemID     v1.SystemID
}

func newServiceClient(c rest.Client, apiServerURL string, systemID v1.SystemID) *ServiceClient {
	return &ServiceClient{
		restClient:   c,
		apiServerURL: apiServerURL,
		systemID:     systemID,
	}
}

func (c *ServiceClient) List() ([]v1.Service, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.ServicesPathFormat, c.systemID))
	body, statusCode, err := c.restClient.Get(url).Body()
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

func (c *ServiceClient) Get(path tree.NodePath) (*v1.Service, error) {
	escapedPath := urlutil.PathEscape(string(path))
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.ServicePathFormat, c.systemID, escapedPath))
	body, statusCode, err := c.restClient.Get(url).Body()
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
