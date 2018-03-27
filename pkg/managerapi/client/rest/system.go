package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	systemSubpath = "/systems"
)

type SystemClient struct {
	restClient rest.Client
	baseURL    string
}

func newSystemClient(c rest.Client, baseURL string) client.SystemClient {
	return &SystemClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, systemSubpath),
	}
}

type createSystemRequest struct {
	ID            types.SystemID `json:"id"`
	DefinitionURL string         `json:"definitionUrl"`
}

func (c *SystemClient) Create(id types.SystemID, definitionURL string) (*types.System, error) {
	request := createSystemRequest{
		ID:            id,
		DefinitionURL: definitionURL,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	system := &types.System{}
	statusCode, err := c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).JSON(&system)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusCreated {
		return system, nil
	}

	if statusCode == http.StatusBadRequest {
		return nil, &client.InvalidSystemOptionsError{}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemClient) List() ([]types.System, error) {
	var systems []types.System
	statusCode, err := c.restClient.Get(c.baseURL).JSON(&systems)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return systems, nil
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemClient) Get(id types.SystemID) (*types.System, error) {
	system := &types.System{}
	statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&system)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return system, nil
	}

	if statusCode == http.StatusNotFound {
		return nil, &client.InvalidSystemIDError{
			ID: id,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemClient) Delete(id types.SystemID) error {
	_, statusCode, err := c.restClient.Delete(fmt.Sprintf("%v/%v", c.baseURL, id)).Body()
	if err != nil {
		return err
	}

	if statusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemClient) SystemBuilds(id types.SystemID) client.SystemBuildClient {
	return newSystemBuildClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id), id)
}

func (c *SystemClient) ServiceBuilds(id types.SystemID) client.ServiceBuildClient {
	return newServiceBuildClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}

func (c *SystemClient) ComponentBuilds(id types.SystemID) client.ComponentBuildClient {
	return newComponentBuildClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id))
}

func (c *SystemClient) Rollouts(id types.SystemID) client.RolloutClient {
	return newRolloutClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id), id)
}

func (c *SystemClient) Teardowns(id types.SystemID) client.TeardownClient {
	return newTeardownClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id), id)
}

func (c *SystemClient) Services(id types.SystemID) client.ServiceClient {
	return newServiceClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id), id)
}

func (c *SystemClient) Secrets(id types.SystemID) client.SystemSecretClient {
	return newSystemSecretClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id), id)
}
