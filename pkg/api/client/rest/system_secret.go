package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	restv1 "github.com/mlab-lattice/system/pkg/api/server/rest/v1"
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
}

func newSystemSecretClient(c rest.Client, baseURL string) *SystemSecretClient {
	return &SystemSecretClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, secretSubpath),
	}
}

func (c *SystemSecretClient) List() ([]v1.Secret, error) {
	body, statusCode, err := c.restClient.Get(c.baseURL).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var secrets []v1.Secret
		err = rest.UnmarshalBodyJSON(body, &secrets)
		return secrets, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *SystemSecretClient) Get(path tree.NodePath, name string) (*v1.Secret, error) {
	secretPath := fmt.Sprintf("%v:%v", path.ToDomain(true), name)
	body, statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, secretPath)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		secret := &v1.Secret{}
		err = rest.UnmarshalBodyJSON(body, &secret)
		return secret, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *SystemSecretClient) Set(path tree.NodePath, name, value string) error {
	secretPath := fmt.Sprintf("%v:%v", path.ToDomain(true), name)

	request := &restv1.SetSecretRequest{
		Value: value,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return err
	}

	body, statusCode, err := c.restClient.PatchJSON(fmt.Sprintf("%v/%v", c.baseURL, secretPath), bytes.NewReader(requestJSON)).Body()
	if err != nil {
		return err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		return nil
	}

	return HandleErrorStatusCode(statusCode, body)
}

func (c *SystemSecretClient) Unset(path tree.NodePath, name string) error {
	secretPath := fmt.Sprintf("%v:%v", path.ToDomain(true), name)

	body, statusCode, err := c.restClient.Delete(fmt.Sprintf("%v/%v", c.baseURL, secretPath)).Body()
	if err != nil {
		return err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		return nil
	}

	return HandleErrorStatusCode(statusCode, body)
}
