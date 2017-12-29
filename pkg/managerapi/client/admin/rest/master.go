package rest

import (
	"fmt"
	"io"

	"github.com/mlab-lattice/system/pkg/managerapi/client/admin"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

const (
	masterNodeSubpath                 = "/master"
	masterNodeComponentSubpath        = "/components"
	masterNodeComponentLogsSubpath    = "/logs"
	masterNodeComponentRestartSubpath = "/restart"
)

type MasterClient struct {
	restClient rest.Client
	baseURL    string
}

func newMasterClient(c rest.Client, baseURL string) admin.MasterClient {
	return &MasterClient{
		restClient: c,
		baseURL:    baseURL + masterNodeSubpath,
	}
}

func (mc *MasterClient) Components() ([]string, error) {
	components := []string{}
	err := mc.restClient.Get(mc.baseURL + masterNodeComponentSubpath).JSON(&components)
	return components, err
}

func (mc *MasterClient) Component(component string) admin.MasterComponentClient {
	return newMasterComponentClient(mc.restClient, mc.baseURL, component)
}

type MasterComponentClient struct {
	restClient rest.Client
	baseURL    string
}

func newMasterComponentClient(c rest.Client, baseURL, component string) admin.MasterComponentClient {
	return &MasterComponentClient{
		restClient: c,
		baseURL:    baseURL + masterNodeComponentSubpath + "/" + component,
	}
}

func (mcc *MasterComponentClient) Logs(nodeID string, follow bool) (io.ReadCloser, error) {
	return mcc.restClient.Get(mcc.baseURL + masterNodeComponentLogsSubpath + fmt.Sprintf("?nodeId=%v&follow=%v", nodeID, follow)).Body()
}

func (mcc *MasterComponentClient) Restart(nodeID string) error {
	_, err := mcc.restClient.Post(mcc.baseURL+masterNodeComponentRestartSubpath+fmt.Sprintf("?nodeId=%v", nodeID), rest.ContentTypeJSON, nil).Do()
	return err
}
