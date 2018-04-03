package versionservice

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/util/rest"
)

type Client interface {
	Status() (bool, error)
	Version() (string, error)
	CheckStatusAndVersion(string) error
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

func (c *DefaultClient) CheckStatusAndVersion(expectedVersion string) error {
	ok, err := c.Status()
	if err != nil {
		return fmt.Errorf("error getting status: %v", err)
	}
	if !ok {
		return fmt.Errorf("status was not okay")
	}

	version, err := c.Version()
	if err != nil {
		return fmt.Errorf("error getting version: %v", err)
	}

	if version != expectedVersion {
		return fmt.Errorf("expected version to be %v but got %v", expectedVersion, version)
	}

	return nil
}
