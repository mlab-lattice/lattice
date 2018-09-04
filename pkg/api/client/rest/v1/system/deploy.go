package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/client/rest/v1/errors"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

type DeployClient struct {
	restClient   rest.Client
	apiServerURL string
	systemID     v1.SystemID
}

func NewDeployClient(c rest.Client, apiServerURL string, systemID v1.SystemID) *DeployClient {
	return &DeployClient{
		restClient:   c,
		apiServerURL: apiServerURL,
		systemID:     systemID,
	}
}

func (c *DeployClient) CreateFromBuild(id v1.BuildID) (*v1.Deploy, error) {
	request := v1rest.DeployRequest{
		BuildID: &id,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.DeploysPathFormat, c.systemID))
	body, statusCode, err := c.restClient.PostJSON(url, bytes.NewReader(requestJSON)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusCreated {
		deploy := &v1.Deploy{}
		err = rest.UnmarshalBodyJSON(body, &deploy)
		return deploy, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *DeployClient) CreateFromVersion(version v1.SystemVersion) (*v1.Deploy, error) {
	request := v1rest.DeployRequest{
		Version: &version,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.DeploysPathFormat, c.systemID))
	body, statusCode, err := c.restClient.PostJSON(url, bytes.NewReader(requestJSON)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusCreated {
		deploy := &v1.Deploy{}
		err = rest.UnmarshalBodyJSON(body, &deploy)
		return deploy, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *DeployClient) List() ([]v1.Deploy, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.DeploysPathFormat, c.systemID))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var deploys []v1.Deploy
		err = rest.UnmarshalBodyJSON(body, &deploys)
		return deploys, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *DeployClient) Get(id v1.DeployID) (*v1.Deploy, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.DeployPathFormat, c.systemID, id))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		deploy := &v1.Deploy{}
		err = rest.UnmarshalBodyJSON(body, &deploy)
		return deploy, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}
