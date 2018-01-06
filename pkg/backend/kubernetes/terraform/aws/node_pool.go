package aws

type NodePool struct {
	Source string `json:"source"`

	Region string `json:"region"`

	ClusterID                 string   `json:"cluster_id"`
	VPCID                     string   `json:"vpc_id"`
	SubnetIDs                 []string `json:"subnet_ids"`
	MasterNodeSecurityGroupID string   `json:"master_node_security_group_id"`
	BaseNodeAMIID             string   `json:"base_node_ami_id"`
	KeyName                   string   `json:"key_name"`

	Name         string `json:"name"`
	NumInstances int32  `json:"num_instances"`
	InstanceType string `json:"instance_type"`
}

func NewNodePoolModule(
	moduleRoot, region, clusterID, vpcID string,
	subnetIDs []string,
	masterNodeSecurityGroupID, baseNodeAMIID, keyName, name string,
	numInstances int32,
	instanceType string,
) *NodePool {
	return &NodePool{
		Source: moduleRoot + modulePathNodePool,

		Region: region,

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
