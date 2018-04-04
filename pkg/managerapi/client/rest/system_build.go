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
	systemBuildSubpath = "/system-builds"
)

type SystemBuildClient struct {
	restClient rest.Client
	baseURL    string
	systemID   types.SystemID
}

func newSystemBuildClient(c rest.Client, baseURL string, systemID types.SystemID) *SystemBuildClient {
	return &SystemBuildClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, systemBuildSubpath),
		systemID:   systemID,
	}
}

type buildSystemRequest struct {
	Version string `json:"version,omitempty"`
}

type buildSystemResponse struct {
	BuildID types.SystemBuildID `json:"buildId"`
}

func (c *SystemBuildClient) Create(version string) (types.SystemBuildID, error) {
	request := &buildSystemRequest{
		Version: version,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	buildResponse := &buildSystemResponse{}
	statusCode, err := c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).JSON(&buildResponse)
	if err != nil {
		return "", err
	}

	if statusCode == http.StatusCreated {
		return buildResponse.BuildID, nil
	}

	if statusCode == http.StatusBadRequest {
		return "", &client.InvalidSystemVersionError{
			Version: version,
		}
	}

	return "", fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemBuildClient) List() ([]types.SystemBuild, error) {
	var builds []types.SystemBuild
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

func (c *SystemBuildClient) Get(id types.SystemBuildID) (*types.SystemBuild, error) {
	build := &types.SystemBuild{}
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
