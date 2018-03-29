package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	v1rest "github.com/mlab-lattice/system/pkg/api/server/rest/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	deploySubpath = "/deploys"
)

type DeployClient struct {
	restClient rest.Client
	baseURL    string
}

func newDeployClient(c rest.Client, baseURL string) *DeployClient {
	return &DeployClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, deploySubpath),
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

	body, statusCode, err := c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusCreated {
		deploy := &v1.Deploy{}
		err = rest.UnmarshalBodyJSON(body, &deploy)
		return deploy, nil
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *DeployClient) CreateFromVersion(version v1.SystemVersion) (*v1.Deploy, error) {
	request := v1rest.DeployRequest{
		Version: &version,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	body, statusCode, err := c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusCreated {
		deploy := &v1.Deploy{}
		err = rest.UnmarshalBodyJSON(body, &deploy)
		return deploy, nil
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *DeployClient) List() ([]v1.Deploy, error) {
	body, statusCode, err := c.restClient.Get(c.baseURL).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var deploys []v1.Deploy
		err = rest.UnmarshalBodyJSON(body, &deploys)
		return deploys, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *DeployClient) Get(id v1.DeployID) (*v1.Deploy, error) {
	body, statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		deploy := &v1.Deploy{}
		err = rest.UnmarshalBodyJSON(body, &deploy)
		return deploy, nil
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}
