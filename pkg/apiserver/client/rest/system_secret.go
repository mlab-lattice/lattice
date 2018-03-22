package rest

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	secretSubpath = "/secrets"
)

type SystemSecretClient struct {
	restClient rest.Client
	baseURL    string
}

func newSystemSecretClient(c rest.Client, baseURL string) *SystemSecretClient {
	return &SystemSecretClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, secretSubpath),
	}
}

func (c *SystemSecretClient) List() ([]types.Secret, error) {
	var secrets []types.Secret
	err := c.restClient.Get(c.baseURL).JSON(&secrets)
	return secrets, err
}

func (c *SystemSecretClient) Get(path tree.NodePath, name string) (*types.Secret, error) {
	secretPath := fmt.Sprintf("%v:%v", path.ToDomain(true), name)
	secret := &types.Secret{}
	err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, secretPath)).JSON(&secret)
	return secret, err
}

type setSystemSecretRequest struct {
	Value string `json:"value"`
}

func (c *SystemSecretClient) Set(path tree.NodePath, name, value string) error {
	secretPath := fmt.Sprintf("%v:%v", path.ToDomain(true), name)

	request := &setSystemSecretRequest{
		Value: value,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return err
	}

	body, err := c.restClient.PatchJSON(fmt.Sprintf("%v/%v", c.baseURL, secretPath), bytes.NewReader(requestJSON)).Body()
	if err != nil {
		return err
	}

	body.Close()
	return nil
}

func (c *SystemSecretClient) Unset(path tree.NodePath, name string) error {
	secretPath := fmt.Sprintf("%v:%v", path.ToDomain(true), name)

	body, err := c.restClient.Delete(fmt.Sprintf("%v/%v", c.baseURL, secretPath)).Body()
	if err != nil {
		return err
	}

	body.Close()
	return nil
}
