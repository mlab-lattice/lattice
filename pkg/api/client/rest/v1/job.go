package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

type JobClient struct {
	restClient   rest.Client
	apiServerURL string
	systemID     v1.SystemID
}

func newJobClient(c rest.Client, apiServerURL string, systemID v1.SystemID) *JobClient {
	return &JobClient{
		restClient:   c,
		apiServerURL: apiServerURL,
		systemID:     systemID,
	}
}

func (c *JobClient) Create(path tree.NodePath, command []string, environment definitionv1.ContainerEnvironment) (*v1.Job, error) {
	request := &v1rest.RunJobRequest{
		Path:        path,
		Command:     command,
		Environment: environment,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.JobsPathFormat, c.systemID))
	body, statusCode, err := c.restClient.PostJSON(url, bytes.NewReader(requestJSON)).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusCreated {
		job := &v1.Job{}
		err = rest.UnmarshalBodyJSON(body, &job)
		return job, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *JobClient) List() ([]v1.Job, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.JobsPathFormat, c.systemID))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var jobs []v1.Job
		err = rest.UnmarshalBodyJSON(body, &jobs)
		return jobs, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *JobClient) Get(id v1.JobID) (*v1.Job, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.JobPathFormat, c.systemID, id))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		job := &v1.Job{}
		err = rest.UnmarshalBodyJSON(body, &job)
		return job, err
	}

	return nil, HandleErrorStatusCode(statusCode, body)
}

func (c *JobClient) Logs(
	id v1.JobID,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	url := fmt.Sprintf(
		"%v%v?%v",
		c.apiServerURL,
		fmt.Sprintf(v1rest.JobLogsPathFormat, c.systemID, id),
		logOptionsToQueryString(logOptions),
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

	return nil, HandleErrorStatusCode(statusCode, body)
}
