package terraform

import (
	"encoding/json"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type NodePool struct {
	Source string

	AWSAccountID string
	Region       string

	LatticeID                 v1.LatticeID
	VPCID                     string
	SubnetIDs                 []string
	MasterNodeSecurityGroupID string
	WorkerNodeAMIID           string
	KeyName                   string

	Name         string
	NumInstances int32
	InstanceType string

	KubeBootstrapToken      string
	LatticeApiServerAddress string
	LatticeApiServerPort    int64
}

func (np *NodePool) MarshalJSON() ([]byte, error) {
	encoder := nodePoolEncoder{
		Source: np.Source,

		AWSAccountID: np.AWSAccountID,
		Region:       np.Region,

		LatticeID:                 string(np.LatticeID),
		VPCID:                     np.VPCID,
		SubnetIDs:                 strings.Join(np.SubnetIDs, ","),
		MasterNodeSecurityGroupID: np.MasterNodeSecurityGroupID,
		WorkerNodeAMIID:           np.WorkerNodeAMIID,
		KeyName:                   np.KeyName,

		Name:         np.Name,
		NumInstances: np.NumInstances,
		InstanceType: np.InstanceType,

		KubeBootstrapToken:      np.KubeBootstrapToken,
		LatticeApiServerAddress: np.LatticeApiServerAddress,
		LatticeApiServerPort:    np.LatticeApiServerPort,
	}
	return json.Marshal(&encoder)
}

type nodePoolEncoder struct {
	Source string `json:"source"`

	AWSAccountID string `json:"aws_account_id"`
	Region       string `json:"region"`

	LatticeID                 string `json:"lattice_id"`
	VPCID                     string `json:"vpc_id"`
	SubnetIDs                 string `json:"subnet_ids"`
	MasterNodeSecurityGroupID string `json:"master_node_security_group_id"`
	WorkerNodeAMIID           string `json:"worker_node_ami_id"`
	KeyName                   string `json:"key_name"`

	Name         string `json:"name"`
	NumInstances int32  `json:"num_instances"`
	InstanceType string `json:"instance_type"`

	KubeBootstrapToken      string `json:"kube_bootstrap_token"`
	LatticeApiServerAddress string `json:"kube_apiserver_address"`
	LatticeApiServerPort    int64  `json:"kube_apiserver_port"`
}
