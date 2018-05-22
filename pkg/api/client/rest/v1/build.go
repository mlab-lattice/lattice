package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	urlutil "net/url"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
	"io"
)

type BuildClient struct {
	restClient   rest.Client
	apiServerURL string
	systemID     v1.SystemID
}

func newBuildClient(c rest.Client, apiServerURL string, systemID v1.SystemID) *BuildClient {
	return &BuildClient{
		restClient:   c,
		apiServerURL: apiServerURL,
		systemID:     systemID,
	}
}

func (c *BuildClient) Create(version v1.SystemVersion) (*v1.Build, error) {
	request := &v1rest.BuildRequest{
		Version: version,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.BuildsPathFormat, c.systemID))
	body, statusCode, err := c.restClient.PostJSON(url, bytes.NewReader(requestJSON)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusCreated {
		build := &v1.Build{}
		err = rest.UnmarshalBodyJSON(body, &build)
		return build, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *BuildClient) List() ([]v1.Build, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.BuildsPathFormat, c.systemID))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var builds []v1.Build
		err = rest.UnmarshalBodyJSON(body, &builds)
		return builds, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *BuildClient) Get(id v1.BuildID) (*v1.Build, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.BuildPathFormat, c.systemID, id))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		build := &v1.Build{}
		err = rest.UnmarshalBodyJSON(body, &build)
		return build, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *BuildClient) Logs(id v1.BuildID, component string, follow bool) (io.ReadCloser, error) {
	escapedPath := urlutil.PathEscape(component)
	url := fmt.Sprintf(
		"%v%v?follow=%v",
		c.apiServerURL,
		fmt.Sprintf(v1rest.BuildLogPathFormat, c.systemID, id, escapedPath),
		follow,
	)
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return body, nil
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}
