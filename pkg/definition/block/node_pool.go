package block

type NodePool struct {
	NumInstances int32  `json:"num_instances"`
	InstanceType string `json:"instance_type"`
}

// Validate implements Interface
func (r *NodePool) Validate(interface{}) error {
	return nil
}
