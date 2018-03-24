package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/server/rest/v1/system"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	systemSubpath = "/systems"
)

type SystemClient struct {
	restClient rest.Client
	baseURL    string
}

func newSystemClient(c rest.Client, baseURL string) clientv1.SystemClient {
	return &SystemClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, systemSubpath),
	}
}

func (c *SystemClient) Create(id v1.SystemID, definitionURL string) (*v1.System, error) {
	request := system.CreateRequest{
		ID:            id,
		DefinitionURL: definitionURL,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	sys := &v1.System{}
	statusCode, err := c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).JSON(&sys)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusCreated {
		return sys, nil
	}

	if statusCode == http.StatusBadRequest {
		return nil, &clientv1.InvalidSystemOptionsError{}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemClient) List() ([]v1.System, error) {
	var systems []v1.System
	statusCode, err := c.restClient.Get(c.baseURL).JSON(&systems)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return systems, nil
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemClient) Get(id v1.SystemID) (*v1.System, error) {
	sys := &v1.System{}
	statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&sys)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return sys, nil
	}

	if statusCode == http.StatusNotFound {
		return nil, &clientv1.InvalidSystemIDError{
			ID: id,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemClient) Delete(id v1.SystemID) error {
	_, statusCode, err := c.restClient.Delete(fmt.Sprintf("%v/%v", c.baseURL, id)).Body()
	if err != nil {
		return err
	}

	if statusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemClient) Builds(id v1.SystemID) clientv1.BuildClient {
	return newBuildClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id), id)
}

func (c *SystemClient) Deploys(id v1.SystemID) clientv1.DeployClient {
	return newDeployClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id), id)
}

func (c *SystemClient) Teardowns(id v1.SystemID) clientv1.TeardownClient {
	return newTeardownClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id), id)
}

func (c *SystemClient) Services(id v1.SystemID) clientv1.ServiceClient {
	return newServiceClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id), id)
}

func (c *SystemClient) Secrets(id v1.SystemID) clientv1.SecretClient {
	return newSystemSecretClient(c.restClient, fmt.Sprintf("%v/%v", c.baseURL, id), id)
}
