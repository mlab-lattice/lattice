package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	ContentTypeJSON = "application/json"

	headerContentType = "Content-Type"
)

type RequestContext struct {
	Client      *http.Client
	Method      string
	Headers     map[string]string
	RequestBody io.Reader
	URL         string
}

func (r *RequestContext) Do() (*http.Response, error) {
	request, err := http.NewRequest(r.Method, r.URL, r.RequestBody)
	if err != nil {
		return nil, err
	}

	return r.Client.Do(request)
}

func (r *RequestContext) Body() (io.ReadCloser, error) {
	response, err := r.Do()
	if err != nil {
		return nil, err
	}

	// FIXME: make this configurable
	if response.StatusCode < 200 || response.StatusCode >= 400 {
		return nil, fmt.Errorf("unexpected status: %v", response.StatusCode)
	}

	return response.Body, nil
}

func (r *RequestContext) JSON(target interface{}) error {
	body, err := r.Body()
	if err != nil {
		return err
	}
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	return json.Unmarshal(bodyBytes, target)
}

type Client interface {
	Get(url string) *RequestContext
	Post(url, contentType string, body io.Reader) *RequestContext
}

func NewClient() *DefaultClient {
	return &DefaultClient{
		client: http.DefaultClient,
	}
}

type DefaultClient struct {
	client *http.Client
}

func (dc *DefaultClient) Get(url string) *RequestContext {
	return &RequestContext{
		Client:      dc.client,
		Method:      http.MethodGet,
		Headers:     nil,
		RequestBody: nil,
		URL:         url,
	}
}

func (dc *DefaultClient) Post(url, contentType string, body io.Reader) *RequestContext {
	return &RequestContext{
		Client: dc.client,
		Method: http.MethodGet,
		Headers: map[string]string{
			headerContentType: contentType,
		},
		RequestBody: body,
		URL:         url,
	}
}
