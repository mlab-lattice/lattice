package system

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/client/rest/v1/errors"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

type TeardownClient struct {
	restClient   rest.Client
	apiServerURL string
	systemID     v1.SystemID
}

func NewTeardownClient(c rest.Client, apiServerURL string, systemID v1.SystemID) *TeardownClient {
	return &TeardownClient{
		restClient:   c,
		apiServerURL: apiServerURL,
		systemID:     systemID,
	}
}

func (c *TeardownClient) Create() (*v1.Teardown, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.TeardownsPathFormat, c.systemID))
	body, statusCode, err := c.restClient.PostJSON(url, nil).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusCreated {
		teardown := &v1.Teardown{}
		err = rest.UnmarshalBodyJSON(body, &teardown)
		return teardown, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *TeardownClient) List() ([]v1.Teardown, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.TeardownsPathFormat, c.systemID))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var teardowns []v1.Teardown
		err = rest.UnmarshalBodyJSON(body, &teardowns)
		return teardowns, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *TeardownClient) Get(id v1.TeardownID) (*v1.Teardown, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.TeardownPathFormat, c.systemID, id))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		teardown := &v1.Teardown{}
		err = rest.UnmarshalBodyJSON(body, &teardown)
		return teardown, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}
