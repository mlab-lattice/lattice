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

	for k, v := range r.Headers {
		request.Header.Set(k, v)
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
	Delete(url string) *RequestContext
	Post(url, contentType string, body io.Reader) *RequestContext
	PostJSON(url string, body io.Reader) *RequestContext
}

func NewClient() *DefaultClient {
	return &DefaultClient{
		defaultHeaders: map[string]string{},
		client:         http.DefaultClient,
	}
}

func NewHeaderedClient(headers map[string]string) *DefaultClient {
	return &DefaultClient{
		defaultHeaders: headers,
		client:         http.DefaultClient,
	}
}

type DefaultClient struct {
	client         *http.Client
	defaultHeaders map[string]string
}

func (dc *DefaultClient) Get(url string) *RequestContext {
	return &RequestContext{
		Client:      dc.client,
		Method:      http.MethodGet,
		Headers:     dc.defaultHeaders,
		RequestBody: nil,
		URL:         url,
	}
}

func (dc *DefaultClient) Post(url, contentType string, body io.Reader) *RequestContext {
	headers := make(map[string]string)
	for k, v := range dc.defaultHeaders {
		headers[k] = v
	}
	headers[headerContentType] = contentType

	return &RequestContext{
		Client:      dc.client,
		Method:      http.MethodPost,
		Headers:     headers,
		RequestBody: body,
		URL:         url,
	}
}

func (dc *DefaultClient) PostJSON(url string, body io.Reader) *RequestContext {
	return dc.Post(url, ContentTypeJSON, body)
}

func (dc *DefaultClient) Delete(url string) *RequestContext {
	return &RequestContext{
		Client:      dc.client,
		Method:      http.MethodDelete,
		Headers:     dc.defaultHeaders,
		RequestBody: nil,
		URL:         url,
	}
}
