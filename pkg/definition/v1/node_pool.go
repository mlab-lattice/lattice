package v1

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type NodePool struct {
	NumInstances int32  `json:"num_instances"`
	InstanceType string `json:"instance_type"`
}

type NodePoolOrReference struct {
	NodePool     *NodePool
	NodePoolPath *tree.NodePathSubcomponent
}

func (np *NodePoolOrReference) MarshalJSON() ([]byte, error) {
	if np.NodePool != nil {
		return json.Marshal(np.NodePool)
	}

	if np.NodePoolPath != nil {
		return json.Marshal(*np.NodePoolPath)
	}

	return json.Marshal(nil)
}

func (np *NodePoolOrReference) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var nodePoolPathString string
	err := json.Unmarshal(data, &nodePoolPathString)
	if err == nil {
		nodePoolPath, err := tree.NewNodePathSubcomponent(nodePoolPathString)
		if err != nil {
			return err
		}

		np.NodePoolPath = &nodePoolPath
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
