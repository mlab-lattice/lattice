package aws

type ServiceDedicatedPrivate struct {
	Source string `json:"source"`

	AWSAccountId string `json:"aws_account_id"`
	Region       string `json:"region"`

	VPCId         string   `json:"vpc_id"`
	SubnetIds     []string `json:"subnet_ids"`
	BaseNodeAmiId string   `json:"base_node_ami_id"`
	KeyName       string   `json:"key_name"`

	SystemId     string `json:"system_id"`
	ServiceId    string `json:"service_id"`
	NumInstances int32  `json:"num_instances"`
	InstanceType string `json:"instance_type"`
}
