package admin

import (
	"fmt"
	"io"
	"net/http"

	"github.com/mlab-lattice/system/pkg/manager/client/requester"
)

const (
	masterNodeSubpath                 = "/master"
	masterNodeComponentSubpath        = "/components"
	masterNodeComponentLogsSubpath    = "/logs"
	masterNodeComponentRestartSubpath = "/restart"
)

type MasterClient struct {
	*Client
}

func newMasterClient(ac *Client) *MasterClient {
	return &MasterClient{
		Client: ac,
	}
}

func (mc *MasterClient) URL(endpoint string) string {
	return mc.Client.URL(masterNodeSubpath + endpoint)
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

func (mnc *MasterNodeClient) URL(endpoint string) string {
	return mnc.MasterClient.URL(fmt.Sprintf("/%v%v", mnc.node, endpoint))
}

func (mnc *MasterNodeClient) Components() ([]string, error) {
	components := []string{}
	err := requester.GetRequestBodyJSON(mnc, masterNodeComponentSubpath, &components)
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

func (mncc *MasterNodeComponentClient) URL(endpoint string) string {
	return mncc.MasterNodeClient.URL(fmt.Sprintf("%v/%v%v", masterNodeComponentSubpath, mncc.component, endpoint))
}

func (mncc *MasterNodeComponentClient) Logs(follow bool) (io.ReadCloser, error) {
	return requester.GetRequestBody(mncc, fmt.Sprintf("%v?follow=%v", masterNodeComponentLogsSubpath, follow))
}

func (mncc *MasterNodeComponentClient) Restart() error {
	resp, err := requester.PostRequestJSON(mncc, masterNodeComponentRestartSubpath, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}
	return nil
}
