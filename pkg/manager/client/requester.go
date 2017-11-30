package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type requester interface {
	httpClient() *http.Client
	url(string) string
}

func makeRequest(requester requester, request *http.Request) (*http.Response, error) {
	return requester.httpClient().Do(request)
}

func getRequest(requester requester, url string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, requester.url(url), nil)
	if err != nil {
		return nil, err
	}
	return makeRequest(requester, request)
}

func postRequest(requester requester, url string, contentType string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPost, requester.url(url), body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", contentType)
	return makeRequest(requester, request)
}

func postRequestJSON(requester requester, url string, body interface{}) (*http.Response, error) {
	var r io.Reader = nil
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		r = bytes.NewReader(buf)
	}
	return postRequest(requester, url, contentTypeApplicationJSON, r)
}

func extractBody(resp *http.Response) (io.ReadCloser, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %v", resp.StatusCode)
	}

	return resp.Body, nil
}

func extractBodyJSON(resp *http.Response, target interface{}) error {
	body, err := extractBody(resp)
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

func getRequestBody(requester requester, url string) (io.ReadCloser, error) {
	resp, err := getRequest(requester, url)
	if err != nil {
		return nil, err
	}
	return extractBody(resp)
}

func getRequestBodyJSON(requester requester, url string, target interface{}) error {
	resp, err := getRequest(requester, url)
	if err != nil {
		return err
	}
	return extractBodyJSON(resp, target)
}
