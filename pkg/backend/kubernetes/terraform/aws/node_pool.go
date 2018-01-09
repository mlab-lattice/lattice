package aws

type NodePool struct {
	Source string `json:"source"`

	AWSAccountID string `json:"aws_account_id"`
	Region       string `json:"region"`

	ClusterID                 string `json:"cluster_id"`
	VPCID                     string `json:"vpc_id"`
	SubnetIDs                 string `json:"subnet_ids"`
	MasterNodeSecurityGroupID string `json:"master_node_security_group_id"`
	BaseNodeAMIID             string `json:"base_node_ami_id"`
	KeyName                   string `json:"key_name"`

	Name         string `json:"name"`
	NumInstances int32  `json:"num_instances"`
	InstanceType string `json:"instance_type"`
}

func NewNodePoolModule(
	moduleRoot, awsAccountID, region, clusterID, vpcID, subnetIDs,
	masterNodeSecurityGroupID, baseNodeAMIID, keyName, name string,
	numInstances int32,
	instanceType string,
) *NodePool {
	return &NodePool{
		Source: moduleRoot + modulePathNodePool,

		AWSAccountID: awsAccountID,
		Region:       region,

		ClusterID:                 clusterID,
		VPCID:                     vpcID,
		SubnetIDs:                 subnetIDs,
		MasterNodeSecurityGroupID: masterNodeSecurityGroupID,
		BaseNodeAMIID:             baseNodeAMIID,
		KeyName:                   keyName,

		Name:         name,
		NumInstances: numInstances,
		InstanceType: instanceType,
	}
}
