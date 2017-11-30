package client

import (
	"fmt"
	"io"
	"net/http"
)

const (
	masterNodeSubpath                 = "/master"
	masterNodeComponentSubpath        = "/components"
	masterNodeComponentLogsSubpath    = "/logs"
	masterNodeComponentRestartSubpath = "/restart"
)

type MasterClient struct {
	client *AdminClient
}

func newMasterClient(client *AdminClient) *MasterClient {
	return &MasterClient{
		client: client,
	}
}

func (mc *MasterClient) Node(node int32) *MasterNodeClient {
	return newMasterNodeClient(mc, node)
}

func (mc *MasterClient) makeRequest(req *http.Request) (*http.Response, error) {
	return mc.client.makeRequest(req)
}

func (mc *MasterClient) url(endpoint string) string {
	return mc.client.url(masterNodeSubpath + endpoint)
}

type MasterNodeClient struct {
	client *MasterClient
	node   int32
}

func newMasterNodeClient(client *MasterClient, node int32) *MasterNodeClient {
	return &MasterNodeClient{
		client: client,
		node:   node,
	}
}

func (mnc *MasterNodeClient) Components() ([]string, error) {
	resp, err := mnc.get(masterNodeComponentSubpath)
	if err != nil {
		return nil, err
	}

	components := []string{}
	err = extractBodyJSON(resp, &components)
	return components, err
}

func (mnc *MasterNodeClient) makeRequest(req *http.Request) (*http.Response, error) {
	return mnc.client.makeRequest(req)
}

func (mnc *MasterNodeClient) get(endpoint string) (*http.Response, error) {
	req, err := getRequest(mnc.url(endpoint))
	if err != nil {
		return nil, err
	}
	return mnc.client.makeRequest(req)
}

func (mnc *MasterNodeClient) url(endpoint string) string {
	return mnc.client.url(fmt.Sprintf("/%v%v", mnc.node, endpoint))
}

func (mnc *MasterNodeClient) Component(component string) *MasterNodeComponentClient {
	return newMasterNodeComponentClient(mnc, component)
}

type MasterNodeComponentClient struct {
	client    *MasterNodeClient
	component string
}

func newMasterNodeComponentClient(client *MasterNodeClient, component string) *MasterNodeComponentClient {
	return &MasterNodeComponentClient{
		client:    client,
		component: component,
	}
}

func (mncc *MasterNodeComponentClient) url(endpoint string) string {
	return mncc.client.url(fmt.Sprintf("%v/%v%v", masterNodeComponentSubpath, mncc.component, endpoint))
}

func (mncc *MasterNodeComponentClient) get(endpoint string) (*http.Response, error) {
	req, err := getRequest(mncc.url(endpoint))
	if err != nil {
		return nil, err
	}
	return mncc.client.makeRequest(req)
}

func (mncc *MasterNodeComponentClient) postJSON(endpoint string, body interface{}) (*http.Response, error) {
	req, err := postRequestJSON(mncc.url(endpoint), body)
	if err != nil {
		return nil, err
	}
	return mncc.client.makeRequest(req)
}

func (mncc *MasterNodeComponentClient) Logs(follow bool) (io.ReadCloser, error) {
	resp, err := mncc.get(fmt.Sprintf("%v?follow=%v", masterNodeComponentLogsSubpath, follow))
	if err != nil {
		return nil, err
	}

	return extractBody(resp)
}

func (mncc *MasterNodeComponentClient) Restart() error {
	_, err := mncc.postJSON(masterNodeComponentRestartSubpath, nil)
	return err
}
