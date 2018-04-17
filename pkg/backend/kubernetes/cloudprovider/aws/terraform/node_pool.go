package terraform

import (
	"strings"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type NodePool struct {
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
}

func NewNodePoolModule(
	moduleRoot, awsAccountID, region string, latticeID v1.LatticeID,
	vpcID string,
	subnetIDs []string,
	masterNodeSecurityGroupID, workerNodeAMIID, keyName, name string,
	numInstances int32,
	instanceType string,
) *NodePool {
	return &NodePool{
		Source: moduleRoot + modulePathNodePool,

		AWSAccountID: awsAccountID,
		Region:       region,

		LatticeID:                 string(latticeID),
		VPCID:                     vpcID,
		SubnetIDs:                 strings.Join(subnetIDs, ","),
		MasterNodeSecurityGroupID: masterNodeSecurityGroupID,
		WorkerNodeAMIID:           workerNodeAMIID,
		KeyName:                   keyName,

		Name:         name,
		NumInstances: numInstances,
		InstanceType: instanceType,
	}
}
