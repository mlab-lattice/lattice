package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	clientv1 "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/server/rest/v1/system"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	secretSubpath = "/secrets"
)

type SystemSecretClient struct {
	restClient rest.Client
	baseURL    string
	systemID   v1.SystemID
}

func newSystemSecretClient(c rest.Client, baseURL string, systemID v1.SystemID) *SystemSecretClient {
	return &SystemSecretClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, secretSubpath),
		systemID:   systemID,
	}
}

func (c *SystemSecretClient) List() ([]v1.Secret, error) {
	var secrets []v1.Secret
	statusCode, err := c.restClient.Get(c.baseURL).JSON(&secrets)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return secrets, err
	}

	if statusCode == http.StatusNotFound {
		return nil, &clientv1.InvalidSystemIDError{
			ID: c.systemID,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemSecretClient) Get(path tree.NodePath, name string) (*v1.Secret, error) {
	secretPath := fmt.Sprintf("%v:%v", path.ToDomain(true), name)
	secret := &v1.Secret{}
	statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, secretPath)).JSON(&secret)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return secret, nil
	}

	if statusCode == http.StatusNotFound {
		// FIXME: need to be able to differentiate between invalid build ID and system ID
		return nil, &clientv1.InvalidSecretError{
			Path: path,
			Name: name,
		}
	}

	return nil, fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemSecretClient) Set(path tree.NodePath, name, value string) error {
	secretPath := fmt.Sprintf("%v:%v", path.ToDomain(true), name)

	request := &system.SetSecretRequest{
		Value: value,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return err
	}

	statusCode, err := c.restClient.PatchJSON(fmt.Sprintf("%v/%v", c.baseURL, secretPath), bytes.NewReader(requestJSON)).Status()
	if err != nil {
		return err
	}

	if statusCode == http.StatusOK {
		return nil
	}

	if statusCode == http.StatusBadRequest {
		return &clientv1.InvalidSecretError{
			Path: path,
			Name: name,
		}
	}

	return fmt.Errorf("unexpected status code %v", statusCode)
}

func (c *SystemSecretClient) Unset(path tree.NodePath, name string) error {
	secretPath := fmt.Sprintf("%v:%v", path.ToDomain(true), name)

	statusCode, err := c.restClient.Delete(fmt.Sprintf("%v/%v", c.baseURL, secretPath)).Status()
	if err != nil {
		return err
	}

	if statusCode == http.StatusOK {
		return nil
	}

	if statusCode == http.StatusBadRequest {
		return &clientv1.InvalidSecretError{
			Path: path,
			Name: name,
		}
	}

	return fmt.Errorf("unexpected status code %v", statusCode)
}
