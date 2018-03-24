package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/apiserver/client"
	"github.com/mlab-lattice/system/pkg/apiserver/server/rest/system"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	deploySubpath = "/deploys"
)

type DeployClient struct {
	restClient rest.Client
	baseURL    string
	systemID   types.SystemID
}

func newDeployClient(c rest.Client, baseURL string, systemID types.SystemID) *DeployClient {
	return &DeployClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, deploySubpath),
		systemID:   systemID,
	}
}

func (c *DeployClient) List() ([]types.Deploy, error) {
	var rollouts []types.Deploy
	statusCode, err := c.restClient.Get(c.baseURL).JSON(&rollouts)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return rollouts, nil
	}

	if statusCode == http.StatusNotFound {
		return nil, &client.InvalidSystemIDError{
			ID: c.systemID,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *DeployClient) Get(id types.DeployID) (*types.Deploy, error) {
	rollout := &types.Deploy{}
	statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&rollout)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return rollout, nil
	}

	if statusCode == http.StatusNotFound {
		return nil, &client.InvalidDeployIDError{
			ID: id,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *DeployClient) CreateFromBuild(id types.BuildID) (types.DeployID, error) {
	request := system.DeployRequest{
		BuildID: &id,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	response := &system.DeployResponse{}
	statusCode, err := c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).JSON(&response)
	if err != nil {
		return "", err
	}

	if statusCode == http.StatusCreated {
		return response.ID, nil
	}

	if statusCode == http.StatusBadRequest {
		return "", &client.InvalidBuildIDError{
			ID: id,
		}
	}

	return "", fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *DeployClient) CreateFromVersion(version string) (types.DeployID, error) {
	request := system.DeployRequest{
		Version: &version,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	response := &system.DeployResponse{}
	statusCode, err := c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).JSON(&response)
	if err != nil {
		return "", err
	}

	if statusCode == http.StatusCreated {
		return response.ID, nil
	}

	if statusCode == http.StatusBadRequest {
		return "", &client.InvalidSystemVersionError{
			Version: version,
		}
	}

	return "", fmt.Errorf("unexpected status code %v", statusCode)
}
