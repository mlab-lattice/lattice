package requester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	contentTypeApplicationJSON = "application/json"
)

type Interface interface {
	HTTPClient() *http.Client
	URL(string) string
}

func MakeRequest(requester Interface, request *http.Request) (*http.Response, error) {
	return requester.HTTPClient().Do(request)
}

func GetRequest(requester Interface, url string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, requester.URL(url), nil)
	if err != nil {
		return nil, err
	}
	return MakeRequest(requester, request)
}

func PostRequest(requester Interface, url string, contentType string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPost, requester.URL(url), body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", contentType)
	return MakeRequest(requester, request)
}

func PostRequestJSON(requester Interface, url string, body interface{}) (*http.Response, error) {
	var r io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		r = bytes.NewReader(buf)
	}
	return PostRequest(requester, url, contentTypeApplicationJSON, r)
}

func ExtractBody(resp *http.Response) (io.ReadCloser, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %v", resp.StatusCode)
	}

	return resp.Body, nil
}

func ExtractBodyJSON(resp *http.Response, target interface{}) error {
	body, err := ExtractBody(resp)
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

func GetRequestBody(requester Interface, url string) (io.ReadCloser, error) {
	resp, err := GetRequest(requester, url)
	if err != nil {
		return nil, err
	}
	return ExtractBody(resp)
}

func GetRequestBodyJSON(requester Interface, url string, target interface{}) error {
	resp, err := GetRequest(requester, url)
	if err != nil {
		return err
	}
	return ExtractBodyJSON(resp, target)
}
