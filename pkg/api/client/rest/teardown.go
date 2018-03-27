package rest

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	teardownSubpath = "/teardowns"
)

type TeardownClient struct {
	restClient rest.Client
	baseURL    string
}

func newTeardownClient(c rest.Client, baseURL string) *TeardownClient {
	return &TeardownClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, teardownSubpath),
	}
}

func (c *TeardownClient) Create() (*v1.Teardown, error) {
	body, statusCode, err := c.restClient.PostJSON(c.baseURL, nil).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusCreated {
		teardown := &v1.Teardown{}
		err = rest.UnmarshalBodyJSON(body, &teardown)
		return teardown, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *TeardownClient) List() ([]v1.Teardown, error) {
	body, statusCode, err := c.restClient.Get(c.baseURL).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var teardowns []v1.Teardown
		err = rest.UnmarshalBodyJSON(body, teardowns)
		return teardowns, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *TeardownClient) Get(id v1.TeardownID) (*v1.Teardown, error) {
	body, statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		teardown := &v1.Teardown{}
		err = rest.UnmarshalBodyJSON(body, &teardown)
		return teardown, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}
