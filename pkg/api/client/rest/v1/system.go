package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/client/rest/v1/errors"
	"github.com/mlab-lattice/lattice/pkg/api/client/rest/v1/system"
	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

type SystemClient struct {
	restClient   rest.Client
	apiServerURL string
}

func newSystemClient(c rest.Client, apiServerURL string) *SystemClient {
	return &SystemClient{
		restClient:   c,
		apiServerURL: apiServerURL,
	}
}

func (c *SystemClient) Create(id v1.SystemID, definitionURL string) (*v1.System, error) {
	request := v1rest.CreateSystemRequest{
		ID:            id,
		DefinitionURL: definitionURL,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	body, statusCode, err := c.restClient.PostJSON(fmt.Sprintf("%v%v", c.apiServerURL, v1rest.SystemsPath), bytes.NewReader(requestJSON)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusCreated {
		system := &v1.System{}
		err = rest.UnmarshalBodyJSON(body, &system)
		return system, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *SystemClient) List() ([]v1.System, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, v1rest.SystemsPath)
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var systems []v1.System
		err = rest.UnmarshalBodyJSON(body, &systems)
		return systems, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *SystemClient) Get(id v1.SystemID) (*v1.System, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.SystemPathFormat, id))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		system := &v1.System{}
		err = rest.UnmarshalBodyJSON(body, system)
		return system, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *SystemClient) Delete(id v1.SystemID) error {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.SystemPathFormat, id))
	body, statusCode, err := c.restClient.Delete(url).Body()
	if err != nil {
		return err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		return nil
	}

	return errors.HandleErrorStatusCode(statusCode, body)
}

func (c *SystemClient) Versions(id v1.SystemID) ([]v1.Version, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.VersionsPathFormat, id))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var versions []v1.Version
		err = rest.UnmarshalBodyJSON(body, &versions)
		return versions, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *SystemClient) Builds(id v1.SystemID) clientv1.SystemBuildClient {
	return system.NewBuildClient(c.restClient, c.apiServerURL, id)
}

func (c *SystemClient) Deploys(id v1.SystemID) clientv1.SystemDeployClient {
	return system.NewDeployClient(c.restClient, c.apiServerURL, id)
}

func (c *SystemClient) Services(id v1.SystemID) clientv1.SystemServiceClient {
	return system.NewServiceClient(c.restClient, c.apiServerURL, id)
}

func (c *SystemClient) Jobs(id v1.SystemID) clientv1.SystemJobClient {
	return system.NewJobClient(c.restClient, c.apiServerURL, id)
}

func (c *SystemClient) Secrets(id v1.SystemID) clientv1.SystemSecretClient {
	return system.NewSecretClient(c.restClient, c.apiServerURL, id)
}

func (c *SystemClient) Teardowns(id v1.SystemID) clientv1.SystemTeardownClient {
	return system.NewTeardownClient(c.restClient, c.apiServerURL, id)
}
