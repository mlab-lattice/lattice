package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/client/rest/v1/errors"
	clientv1 "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

type JobClient struct {
	restClient   rest.Client
	apiServerURL string
	system       v1.SystemID
}

func NewJobClient(c rest.Client, apiServerURL string, systemID v1.SystemID) *JobClient {
	return &JobClient{
		restClient:   c,
		apiServerURL: apiServerURL,
		system:       systemID,
	}
}

func (c *JobClient) Run(
	path tree.Path,
	command []string,
	environment definitionv1.ContainerExecEnvironment,
	numRetries *int32,
) (*v1.Job, error) {
	request := &v1rest.RunJobRequest{
		Path:        path,
		Command:     command,
		Environment: environment,
		NumRetries:  numRetries,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.JobsPathFormat, c.system))
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

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *JobClient) List() ([]v1.Job, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.JobsPathFormat, c.system))
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

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *JobClient) Get(id v1.JobID) (*v1.Job, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.JobPathFormat, c.system, id))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		job := new(v1.Job)
		err = rest.UnmarshalBodyJSON(body, &job)
		return job, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *JobClient) Runs(id v1.JobID) clientv1.SystemJobRunClient {
	return NewJobRunClient(c.restClient, c.apiServerURL, c.system, id)
}

type JobRunClient struct {
	restClient   rest.Client
	apiServerURL string
	system       v1.SystemID
	job          v1.JobID
}

func NewJobRunClient(c rest.Client, apiServerURL string, system v1.SystemID, job v1.JobID) *JobRunClient {
	return &JobRunClient{
		restClient:   c,
		apiServerURL: apiServerURL,
		system:       system,
		job:          job,
	}
}

func (c *JobRunClient) List() ([]v1.JobRun, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.JobRunsPathFormat, c.system, c.job))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode == http.StatusOK {
		var runs []v1.JobRun
		err = rest.UnmarshalBodyJSON(body, &runs)
		return runs, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *JobRunClient) Get(id v1.JobRunID) (*v1.JobRun, error) {
	url := fmt.Sprintf("%v%v", c.apiServerURL, fmt.Sprintf(v1rest.JobRunPathFormat, c.system, c.job, id))
	body, statusCode, err := c.restClient.Get(url).Body()
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusOK {
		run := new(v1.JobRun)
		err = rest.UnmarshalBodyJSON(body, &run)
		return run, err
	}

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}

func (c *JobRunClient) Logs(
	id v1.JobRunID,
	sidecar *string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	url := fmt.Sprintf(
		"%v%v?%v",
		c.apiServerURL,
		fmt.Sprintf(v1rest.JobRunLogsPathFormat, c.system, c.job, id),
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

	return nil, errors.HandleErrorStatusCode(statusCode, body)
}
