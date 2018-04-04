package aws

type Lattice struct {
	Source string `json:"source"`

	AWSAccountID string `json:"aws_account_id"`
	Region       string `json:"region"`

	AvailabilityZones []string `json:"availability_zones"`

	LatticeID                    string `json:"lattice_id"`
	ControlPlaneContainerChannel string `json:"control_plane_container_channel"`
	SystemDefinitionURL          string `json:"system_definition_url"`

	BaseNodeAMIID          string `json:"base_node_ami_id"`
	MasterNodeAMIID        string `json:"master_node_ami_id"`
	MasterNodeInstanceType string `json:"master_node_instance_type"`
	KeyName                string `json:"key_name"`

	APIServerPort int32 `json:"api_server_port"`
}
