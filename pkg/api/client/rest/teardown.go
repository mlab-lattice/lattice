package rest

import (
	"fmt"
	"net/http"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/server/rest/v1/system"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	teardownSubpath = "/teardowns"
)

type TeardownClient struct {
	restClient rest.Client
	baseURL    string
	systemID   v1.SystemID
}

func newTeardownClient(c rest.Client, baseURL string, systemID v1.SystemID) *TeardownClient {
	return &TeardownClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, teardownSubpath),
		systemID:   systemID,
	}
}

func (c *TeardownClient) List() ([]v1.SystemTeardown, error) {
	var teardowns []v1.SystemTeardown
	statusCode, err := c.restClient.Get(c.baseURL).JSON(&teardowns)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return teardowns, nil
	}

	if statusCode == http.StatusNotFound {
		return nil, &clientv1.InvalidSystemIDError{
			ID: c.systemID,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *TeardownClient) Get(id v1.TeardownID) (*v1.SystemTeardown, error) {
	teardown := &v1.SystemTeardown{}
	statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&teardown)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return teardown, nil
	}

	if statusCode == http.StatusNotFound {
		return nil, &clientv1.InvalidTeardownIDError{
			ID: id,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *TeardownClient) Create() (v1.TeardownID, error) {
	teardownResponse := &system.TearDownResponse{}
	statusCode, err := c.restClient.PostJSON(c.baseURL, nil).JSON(&teardownResponse)
	if err != nil {
		return "", err
	}

	if statusCode == http.StatusCreated {
		return teardownResponse.ID, nil
	}

	if statusCode == http.StatusBadRequest {
		return "", &clientv1.InvalidSystemIDError{
			ID: c.systemID,
		}
	}

	return "", fmt.Errorf("unexpected status code %v", statusCode)
}
