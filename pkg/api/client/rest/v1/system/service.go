package system

import (
	"fmt"
	"io"
	"net/http"
	urlutil "net/url"

	"github.com/mlab-lattice/lattice/pkg/api/client/rest/v1/errors"
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

func NewServiceClient(c rest.Client, apiServerURL string, systemID v1.SystemID) *ServiceClient {
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

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *ServiceClient) Get(id v1.ServiceID) (*v1.Service, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.ServicePathFormat, c.systemID, id))
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

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *ServiceClient) GetByServicePath(path tree.Path) (*v1.Service, error) {
	escapedPath := urlutil.PathEscape(path.String())
	url := fmt.Sprintf("%v%v?path=%v", c.apiServerURL,
		fmt.Sprintf(v1rest.ServicesPathFormat, c.systemID), escapedPath)
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var services []v1.Service
		err = rest.UnmarshalBodyJSON(body, &services)

		if err != nil {
			return nil, err
		}

		if len(services) != 1 {
			return nil, fmt.Errorf("server returned more than one service for path '%s'", path)
		}
		return &services[0], err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *ServiceClient) Logs(
	id v1.ServiceID,
	sidecar, instance *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	url := fmt.Sprintf(
		"%v%v?%v",
		c.apiServerURL,
		fmt.Sprintf(v1rest.ServiceLogsPathFormat, c.systemID, id),
		logOptionsToQueryString(logOptions),
	)

	if sidecar != nil {
		url += fmt.Sprintf("&sidecar=%v", *sidecar)
	}

	if instance != nil {
		url += fmt.Sprintf("&instance=%v", *instance)
	}

	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return body, nil
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}
