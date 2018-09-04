package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	urlutil "net/url"

	"github.com/mlab-lattice/lattice/pkg/api/client/rest/v1/errors"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

type SecretClient struct {
	restClient   rest.Client
	apiServerURL string
	systemID     v1.SystemID
}

func NewSecretClient(c rest.Client, apiServerURL string, systemID v1.SystemID) *SecretClient {
	return &SecretClient{
		restClient:   c,
		apiServerURL: apiServerURL,
		systemID:     systemID,
	}
}

func (c *SecretClient) List() ([]v1.Secret, error) {
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

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *SecretClient) Get(path tree.PathSubcomponent) (*v1.Secret, error) {
	escapedPath := urlutil.PathEscape(path.String())
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

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *SecretClient) Set(path tree.PathSubcomponent, value string) error {
	request := &v1rest.SetSecretRequest{
		Value: value,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return err
	}

	escapedPath := urlutil.PathEscape(path.String())
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.SystemSecretPathFormat, c.systemID, escapedPath))
	body, statusCode, err := c.restClient.PatchJSON(url, bytes.NewReader(requestJSON)).Body()
	if err != nil {
		return err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		return nil
	}

	return errors.HandleErrorStatusCode(statusCode, body)
}

func (c *SecretClient) Unset(path tree.PathSubcomponent) error {
	escapedPath := urlutil.PathEscape(string(path))
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.SystemSecretPathFormat, c.systemID, escapedPath))
	body, statusCode, err := c.restClient.Delete(url).Body()
	if err != nil {
		return err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		return nil
	}

	return errors.HandleErrorStatusCode(statusCode, body)
}
