package rest

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	teardownSubpath = "/teardowns"
)

type TeardownClient struct {
	restClient rest.Client
	baseURL    string
	systemID   types.SystemID
}

func newTeardownClient(c rest.Client, baseURL string, systemID types.SystemID) *TeardownClient {
	return &TeardownClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, teardownSubpath),
		systemID:   systemID,
	}
}

func (c *TeardownClient) List() ([]types.SystemTeardown, error) {
	var teardowns []types.SystemTeardown
	statusCode, err := c.restClient.Get(c.baseURL).JSON(&teardowns)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return teardowns, nil
	}

	if statusCode == http.StatusNotFound {
		return nil, &client.InvalidSystemIDError{
			ID: c.systemID,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *TeardownClient) Get(id types.SystemTeardownID) (*types.SystemTeardown, error) {
	teardown := &types.SystemTeardown{}
	statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&teardown)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return teardown, nil
	}

	if statusCode == http.StatusNotFound {
		return nil, &client.InvalidTeardownIDError{
			ID: id,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

type teardownResponse struct {
	TeardownID types.SystemTeardownID `json:"teardownId"`
}

func (c *TeardownClient) Create() (types.SystemTeardownID, error) {
	teardownResponse := &teardownResponse{}
	statusCode, err := c.restClient.PostJSON(c.baseURL, nil).JSON(&teardownResponse)
	if err != nil {
		return "", err
	}

	if statusCode == http.StatusCreated {
		return teardownResponse.TeardownID, nil
	}

	if statusCode == http.StatusBadRequest {
		return "", &client.InvalidSystemIDError{
			ID: c.systemID,
		}
	}

	return "", fmt.Errorf("unexpected status code %v", statusCode)
}
