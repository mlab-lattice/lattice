package aws

type ServiceDedicatedPrivate struct {
	Source string `json:"source"`

	AWSAccountID string `json:"aws_account_id"`
	Region       string `json:"region"`

	VPCID                     string `json:"vpc_id"`
	SubnetIDs                 string `json:"subnet_ids"`
	MasterNodeSecurityGroupID string `json:"master_node_security_group_id"`
	BaseNodeAmiID             string `json:"base_node_ami_id"`
	KeyName                   string `json:"key_name"`

	SystemID     string `json:"system_id"`
	ServiceID    string `json:"service_id"`
	NumInstances int32  `json:"num_instances"`
	InstanceType string `json:"instance_type"`
}

type ServiceDedicatedPublicHTTP struct {
	Source string `json:"source"`

	AWSAccountID string `json:"aws_account_id"`
	Region       string `json:"region"`

	VPCID                     string `json:"vpc_id"`
	SubnetIDs                 string `json:"subnet_ids"`
	MasterNodeSecurityGroupID string `json:"master_node_security_group_id"`
	BaseNodeAmiID             string `json:"base_node_ami_id"`
	KeyName                   string `json:"key_name"`

	SystemID     string `json:"system_id"`
	ServiceID    string `json:"service_id"`
	NumInstances int32  `json:"num_instances"`
	InstanceType string `json:"instance_type"`

	Ports map[int32]int32 `json:"ports"`
}
