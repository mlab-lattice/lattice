package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	v1rest "github.com/mlab-lattice/system/pkg/api/server/rest/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	systemBuildSubpath = "/builds"
)

type BuildClient struct {
	restClient rest.Client
	baseURL    string
}

func newBuildClient(c rest.Client, baseURL string) *BuildClient {
	return &BuildClient{
		restClient: c,
		baseURL:    fmt.Sprintf("%v%v", baseURL, systemBuildSubpath),
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

	body, statusCode, err := c.restClient.PostJSON(c.baseURL, bytes.NewReader(requestJSON)).Body()
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
	body, statusCode, err := c.restClient.Get(c.baseURL).Body()
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
	body, statusCode, err := c.restClient.Get(fmt.Sprintf("%v/%v", c.baseURL, id)).Body()
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
