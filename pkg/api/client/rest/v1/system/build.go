package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	urlutil "net/url"

	"io"

	"github.com/mlab-lattice/lattice/pkg/api/client/rest/v1/errors"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

type BuildClient struct {
	restClient   rest.Client
	apiServerURL string
	systemID     v1.SystemID
}

func NewBuildClient(c rest.Client, apiServerURL string, systemID v1.SystemID) *BuildClient {
	return &BuildClient{
		restClient:   c,
		apiServerURL: apiServerURL,
		systemID:     systemID,
	}
}

func (c *BuildClient) CreateFromPath(path tree.Path) (*v1.Build, error) {
	return c.create(&path, nil)
}

func (c *BuildClient) CreateFromVersion(version v1.Version) (*v1.Build, error) {
	return c.create(nil, &version)
}

func (c *BuildClient) create(path *tree.Path, version *v1.Version) (*v1.Build, error) {
	request := &v1rest.BuildRequest{
		Path:    path,
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

	return nil, errors.HandleErrorStatusCode(statusCode, body)
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

	return nil, errors.HandleErrorStatusCode(statusCode, body)
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

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *BuildClient) Logs(
	id v1.BuildID,
	path tree.Path,
	sidecar *string,
	options *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	escapedPath := urlutil.PathEscape(path.String())
	url := fmt.Sprintf(
		"%v%v?path=%v&%v",
		c.apiServerURL,
		fmt.Sprintf(v1rest.BuildLogsPathFormat, c.systemID, id),
		escapedPath,
		logOptionsToQueryString(options),
	)

	if sidecar != nil {
		url += fmt.Sprintf("&sidecar=%v", *sidecar)
	}

	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		return body, nil
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}
