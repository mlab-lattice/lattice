package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	urlutil "net/url"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

type SystemSecretClient struct {
	restClient   rest.Client
	apiServerURL string
	systemID     v1.SystemID
}

func newSystemSecretClient(c rest.Client, apiServerURL string, systemID v1.SystemID) *SystemSecretClient {
	return &SystemSecretClient{
		restClient:   c,
		apiServerURL: apiServerURL,
		systemID:     systemID,
	}
}

func (c *SystemSecretClient) List() ([]v1.Secret, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.SystemSecretsPathFormat, c.systemID))
	body, statusCode, err := c.restClient.Get(url).Body()
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
	escapedPath := fmt.Sprintf("%v:%v", urlutil.PathEscape(string(path)), name)
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.SystemSecretPathFormat, c.systemID, escapedPath))
	body, statusCode, err := c.restClient.Get(url).Body()
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
	request := &v1rest.SetSecretRequest{
		Value: value,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return err
	}

	escapedPath := fmt.Sprintf("%v:%v", urlutil.PathEscape(string(path)), name)
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.SystemSecretPathFormat, c.systemID, escapedPath))
	body, statusCode, err := c.restClient.PatchJSON(url, bytes.NewReader(requestJSON)).Body()
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
	escapedPath := fmt.Sprintf("%v:%v", urlutil.PathEscape(string(path)), name)
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.SystemSecretPathFormat, c.systemID, escapedPath))
	body, statusCode, err := c.restClient.Delete(url).Body()
	if err != nil {
		return err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		return nil
	}

	return HandleErrorStatusCode(statusCode, body)
}
