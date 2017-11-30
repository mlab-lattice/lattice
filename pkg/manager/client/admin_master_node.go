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
	*AdminClient
}

func newMasterClient(ac *AdminClient) *MasterClient {
	return &MasterClient{
		AdminClient: ac,
	}
}

func (mc *MasterClient) url(endpoint string) string {
	return mc.AdminClient.url(masterNodeSubpath + endpoint)
}

func (mc *MasterClient) Node(node int32) *MasterNodeClient {
	return newMasterNodeClient(mc, node)
}

type MasterNodeClient struct {
	*MasterClient
	node int32
}

func newMasterNodeClient(mc *MasterClient, node int32) *MasterNodeClient {
	return &MasterNodeClient{
		MasterClient: mc,
		node:         node,
	}
}

func (mnc *MasterNodeClient) url(endpoint string) string {
	return mnc.MasterClient.url(fmt.Sprintf("/%v%v", mnc.node, endpoint))
}

func (mnc *MasterNodeClient) Components() ([]string, error) {
	components := []string{}
	err := getRequestBodyJSON(mnc, masterNodeComponentSubpath, &components)
	return components, err
}

func (mnc *MasterNodeClient) Component(component string) *MasterNodeComponentClient {
	return newMasterNodeComponentClient(mnc, component)
}

type MasterNodeComponentClient struct {
	*MasterNodeClient
	component string
}

func newMasterNodeComponentClient(mnc *MasterNodeClient, component string) *MasterNodeComponentClient {
	return &MasterNodeComponentClient{
		MasterNodeClient: mnc,
		component:        component,
	}
}

func (mncc *MasterNodeComponentClient) url(endpoint string) string {
	return mncc.MasterNodeClient.url(fmt.Sprintf("%v/%v%v", masterNodeComponentSubpath, mncc.component, endpoint))
}

func (mncc *MasterNodeComponentClient) Logs(follow bool) (io.ReadCloser, error) {
	return getRequestBody(mncc, fmt.Sprintf("%v?follow=%v", masterNodeComponentLogsSubpath, follow))
}

func (mncc *MasterNodeComponentClient) Restart() error {
	resp, err := postRequestJSON(mncc, masterNodeComponentRestartSubpath, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}
	return nil
}
