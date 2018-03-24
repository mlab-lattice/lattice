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
	deploySubpath = "/deploys"
)

type DeployClient struct {
	restClient rest.Client
	baseURL    string
	systemID   v1.SystemID
}

func newDeployClient(c rest.Client, baseURL string, systemID v1.SystemID) *DeployClient {
	return &DeployClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, deploySubpath),
		systemID:   systemID,
	}
}

func (c *DeployClient) List() ([]v1.Deploy, error) {
	var rollouts []v1.Deploy
	statusCode, err := c.restClient.Get(c.baseURL).JSON(&rollouts)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return rollouts, nil
	}

	if statusCode == http.StatusNotFound {
		return nil, &clientv1.InvalidSystemIDError{
			ID: c.systemID,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *DeployClient) Get(id v1.DeployID) (*v1.Deploy, error) {
	rollout := &v1.Deploy{}
	statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&rollout)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return rollout, nil
	}

	if statusCode == http.StatusNotFound {
		return nil, &clientv1.InvalidDeployIDError{
			ID: id,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *DeployClient) CreateFromBuild(id v1.BuildID) (v1.DeployID, error) {
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
		return "", &clientv1.InvalidBuildIDError{
			ID: id,
		}
	}

	return "", fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *DeployClient) CreateFromVersion(version string) (v1.DeployID, error) {
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
		return "", &clientv1.InvalidSystemVersionError{
			Version: version,
		}
	}

	return "", fmt.Errorf("unexpected status code %v", statusCode)
}
