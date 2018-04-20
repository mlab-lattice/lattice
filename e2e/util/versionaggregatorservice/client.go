package versionaggregatorservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/mlab-lattice/lattice/pkg/util/rest"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type Aggregation struct {
	VersionServices           []VersionServiceResponseInfo           `json:"versionServices"`
	VersionAggregatorServices []VersionAggregatorServiceResponseInfo `json:"versionAggregatorServices"`
}

type VersionServiceResponseInfo struct {
	URL    string                      `json:"url"`
	Status *int                        `json:"status,omitempty"`
	Body   *VersionServiceResponseBody `json:"body,omitempty"`
	Error  *string                     `json:"error,omitempty"`
}

type VersionServiceResponseBody struct {
	Version string `json:"version"`
}

type VersionAggregatorServiceResponseInfo struct {
	URL    string       `json:"url"`
	Status *int         `json:"status,omitempty"`
	Body   *Aggregation `json:"body,omitempty"`
	Error  *string      `json:"error,omitempty"`
}

type RequestBody struct {
	VersionServices           []VersionService           `json:"versionServices,omitempty"`
	VersionAggregatorServices []VersionAggregatorService `json:"versionAggregatorServices,omitempty"`
}

type VersionService struct {
	URL string `json:"url"`
}

type VersionAggregatorService struct {
	URL         string       `json:"url"`
	RequestBody *RequestBody `json:"requestBody"`
}

type Client interface {
	Status() (bool, error)
	Aggregate([]VersionService, []VersionAggregatorService) (*Aggregation, error)
}

func NewClient(url string) *DefaultClient {
	return &DefaultClient{
		url:    url,
		client: rest.NewClient(),
	}
}

type DefaultClient struct {
	url    string
	client rest.Client
}

type statusResponse struct {
	OK bool `json:"ok"`
}

func (c *DefaultClient) Status() (bool, error) {
	response := statusResponse{}
	status, err := c.client.Get(fmt.Sprintf("%v/status", c.url)).JSON(&response)
	if err != nil {
		return false, err
	}

	if status != http.StatusOK {
		return false, fmt.Errorf("unexpected status: %v", status)
	}

	return response.OK, nil
}

func (c *DefaultClient) Aggregate(versionServices []VersionService, aggregatorServices []VersionAggregatorService) (*Aggregation, error) {
	requestBody := &RequestBody{
		VersionServices:           versionServices,
		VersionAggregatorServices: aggregatorServices,
	}

	requestJSON, err := json.Marshal(&requestBody)
	if err != nil {
		return nil, err
	}

	response := &Aggregation{}
	status, err := c.client.PostJSON(fmt.Sprintf("%v/aggregate", c.url), bytes.NewReader(requestJSON)).JSON(&response)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %v", status)
	}

	return response, nil
}

func (c *DefaultClient) CheckStatusAndAggregation(versionServices []VersionService, aggregatorServices []VersionAggregatorService, expectedAggregation *Aggregation) error {
	ok, err := c.Status()
	if err != nil {
		return fmt.Errorf("error getting status: %v", err)
	}
	if !ok {
		return fmt.Errorf("status was not okay")
	}

	aggregation, err := c.Aggregate(versionServices, aggregatorServices)
	if err != nil {
		return fmt.Errorf("error getting aggregation: %v", err)
	}

	if !reflect.DeepEqual(expectedAggregation, aggregation) {
		expectedJSON, err := json.Marshal(expectedAggregation)
		if err != nil {
			return err
		}

		aggregationJSON, err := json.Marshal(aggregation)
		if err != nil {
			return err
		}

		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(expectedJSON), string(aggregationJSON), true)
		return fmt.Errorf("expected aggregation did not match the result: %v", dmp.DiffPrettyText(diffs))
	}

	return nil
}
