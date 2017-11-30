package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	adminEndpontPath           = "/admin"
	contentTypeApplicationJSON = "application/json"
)

type AdminClient struct {
	httpClient    *http.Client
	managerAPIURL string
}

func NewClient(managerAPIURL string) *AdminClient {
	return &AdminClient{
		httpClient:    http.DefaultClient,
		managerAPIURL: managerAPIURL,
	}
}

func (ac *AdminClient) makeRequest(req *http.Request) (*http.Response, error) {
	return ac.httpClient.Do(req)
}

func (ac *AdminClient) url(endpoint string) string {
	return ac.managerAPIURL + adminEndpontPath + endpoint
}

func (ac *AdminClient) Master() *MasterClient {
	return newMasterClient(ac)
}

func getRequest(url string) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, url, nil)
}

func postRequest(url string, contentType string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return req, nil
}

func postRequestJSON(url string, body interface{}) (*http.Request, error) {
	var r io.Reader = nil
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		r = bytes.NewReader(buf)
	}
	return postRequest(url, contentTypeApplicationJSON, r)
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
