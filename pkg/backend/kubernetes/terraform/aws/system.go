package aws

type System struct {
	Source string `json:"source"`

	AWSAccountID      string   `json:"aws_account_id"`
	Region            string   `json:"region"`
	AvailabilityZones []string `json:"availability_zones"`
	KeyName           string   `json:"key_name"`

	SystemID            string `json:"system_id"`
	SystemDefinitionURL string `json:"system_definition_url"`

	MasterNodeInstanceType          string `json:"master_node_instance_type"`
	MasterNodeAMIID                 string `json:"master_node_ami_id"`
	SystemEnvironmentManagerAPIPort int32  `json:"system_environment_manager_api_port"`

	BaseNodeAMIID string `json:"base_node_ami_id"`
}
