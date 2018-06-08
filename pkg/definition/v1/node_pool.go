package v1

import (
	"encoding/json"
	"fmt"
)

type NodePool struct {
	NumInstances int32  `json:"num_instances"`
	InstanceType string `json:"instance_type"`
}

type NodePoolOrReference struct {
	NodePool     *NodePool
	NodePoolName *string
}

func (np *NodePoolOrReference) MarshalJSON() ([]byte, error) {
	if np.NodePool != nil {
		return json.Marshal(np.NodePool)
	}

	if np.NodePoolName != nil {
		return json.Marshal(*np.NodePoolName)
	}

	return json.Marshal(nil)
}

func (np *NodePoolOrReference) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var nodePoolName string
	err := json.Unmarshal(data, &nodePoolName)
	if err == nil {
		np.NodePoolName = &nodePoolName
		return nil
	}

	// Ensure the Unmarshalling failed due to the data not being a string
	if _, ok := err.(*json.UnmarshalTypeError); !ok {
		return err
	}

	var nodePool *NodePool
	err = json.Unmarshal(data, &nodePool)
	if err == nil {
		np.NodePool = nodePool
		return nil
	}

	// Ensure the Unmarshalling failed due to the data not being a string
	if _, ok := err.(*json.UnmarshalTypeError); !ok {
		return err
	}

	return fmt.Errorf("invalid node_pool json")
}
