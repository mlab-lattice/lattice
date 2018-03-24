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
	rolloutSubpath = "/rollouts"
)

type RolloutClient struct {
	restClient rest.Client
	baseURL    string
	systemID   types.SystemID
}

func newRolloutClient(c rest.Client, baseURL string, systemID types.SystemID) *RolloutClient {
	return &RolloutClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, rolloutSubpath),
		systemID:   systemID,
	}
}

func (c *RolloutClient) List() ([]types.SystemRollout, error) {
	var rollouts []types.SystemRollout
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

func (c *RolloutClient) Get(id types.SystemRolloutID) (*types.SystemRollout, error) {
	rollout := &types.SystemRollout{}
	statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).JSON(&rollout)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return rollout, nil
	}

	if statusCode == http.StatusNotFound {
		return nil, &client.InvalidRolloutIDError{
			ID: id,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

type rollOutSystemRequest struct {
	Version *string              `json:"version,omitempty"`
	BuildID *types.SystemBuildID `json:"buildId,omitempty"`
}

type rolloutResponse struct {
	RolloutID types.SystemRolloutID `json:"rolloutId"`
}

func (c *RolloutClient) CreateFromBuild(id types.SystemBuildID) (types.SystemRolloutID, error) {
	request := rollOutSystemRequest{
		BuildID: &id,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	rolloutResponse := &rolloutResponse{}
	statusCode, err := c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).JSON(&rolloutResponse)
	if err != nil {
		return "", err
	}

	if statusCode == http.StatusCreated {
		return rolloutResponse.RolloutID, nil
	}

	if statusCode == http.StatusBadRequest {
		return "", &client.InvalidBuildIDError{
			ID: id,
		}
	}

	return "", fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *RolloutClient) CreateFromVersion(version string) (types.SystemRolloutID, error) {
	request := rollOutSystemRequest{
		Version: &version,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	rolloutResponse := &rolloutResponse{}
	statusCode, err := c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).JSON(&rolloutResponse)
	if err != nil {
		return "", err
	}

	if statusCode == http.StatusCreated {
		return rolloutResponse.RolloutID, nil
	}

	if statusCode == http.StatusBadRequest {
		return "", &client.InvalidSystemVersionError{
			Version: version,
		}
	}

	return "", fmt.Errorf("unexpected status code %v", statusCode)
}
