package versionservice

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/util/rest"
)

type Client interface {
	Status() (bool, error)
	Version() (string, error)
}

func NewClient(url string) *DefaultClient {
	return &DefaultClient{
		url:    url,
		client: rest.NewClient(),
	}
}

type DefaultClient struct {
	url    string
	client rest.Client
}

type statusResponse struct {
	OK bool `json:"ok"`
}

func (c *DefaultClient) Status() (bool, error) {
	response := statusResponse{}
	status, err := c.client.Get(fmt.Sprintf("%v/status", c.url)).JSON(&response)
	if err != nil {
		return false, err
	}

	if status != http.StatusOK {
		return false, fmt.Errorf("unexpected status: %v", status)
	}

	return response.OK, nil
}

type versionResponse struct {
	Version string `json:"version"`
}

func (c *DefaultClient) Version() (string, error) {
	response := versionResponse{}
	status, err := c.client.Get(fmt.Sprintf("%v/version", c.url)).JSON(&response)
	if err != nil {
		return "", err
	}

	if status != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %v", status)
	}

	return response.Version, nil
}
