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
	systemBuildSubpath = "/builds"
)

type BuildClient struct {
	restClient rest.Client
	baseURL    string
	systemID   types.SystemID
}

func newBuildClient(c rest.Client, baseURL string, systemID types.SystemID) *BuildClient {
	return &BuildClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, systemBuildSubpath),
		systemID:   systemID,
	}
}

func (c *BuildClient) Create(version string) (types.BuildID, error) {
	request := &system.BuildRequest{
		Version: version,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	response := &system.BuildResponse{}
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

func (c *BuildClient) List() ([]types.Build, error) {
	var builds []types.Build
	statusCode, err := c.restClient.Get(c.baseURL).JSON(&builds)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return builds, err
	}

	if statusCode == http.StatusNotFound {
		return nil, &client.InvalidSystemIDError{
			ID: c.systemID,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *BuildClient) Get(id types.BuildID) (*types.Build, error) {
	build := &types.Build{}
	statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&build)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return build, nil
	}

	if statusCode == http.StatusNotFound {
		// FIXME: need to be able to differentiate between invalid build ID and system ID
		return nil, &client.InvalidBuildIDError{
			ID: id,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}
