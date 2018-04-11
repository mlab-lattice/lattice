package block

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Resources struct {
	// TODO: add scaling
	MinInstances *int32             `json:"min_instances,omitempty"`
	MaxInstances *int32             `json:"max_instances,omitempty"`
	NumInstances *int32             `json:"num_instances,omitempty"`
	InstanceType *string            `json:"instance_type,omitempty"`
	NodePool     *ResourcesNodePool `json:"node_pool,omitempty"`
}

type ResourcesNodePool struct {
	NodePool     *NodePool
	NodePoolName *string
}

func (np *ResourcesNodePool) MarshalJSON() ([]byte, error) {
	if np.NodePool != nil {
		return json.Marshal(np.NodePool)
	}

	if np.NodePoolName != nil {
		return json.Marshal(*np.NodePoolName)
	}

	return json.Marshal(nil)
}

func (np *ResourcesNodePool) UnmarshalJSON(data []byte) error {
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

// Validate implements Interface
func (r *Resources) Validate(interface{}) error {
	if r.MinInstances == nil && r.MaxInstances == nil && r.NumInstances == nil {
		return errors.New("must set either num_instances or min_instances and max_instances")
	}

	if r.NumInstances != nil {
		if r.MinInstances != nil || r.MaxInstances != nil {
			return errors.New("can only set either num_instances or min_instances and max_instances")
		}

		if *r.NumInstances < 1 {
			return errors.New("invalid num_instances")
		}
	} else {
		if r.MinInstances == nil || r.MaxInstances == nil {
			return errors.New("must set either num_instances or min_instances and max_instances")
		}

		if *r.MinInstances < 1 {
			return errors.New("invalid min_instances")
		}

		if *r.MaxInstances < *r.MinInstances {
			return errors.New("max_instances must be greater than or equal to min_instances")
		}
	}

	// TODO: cap max instances
	// TODO: conditionally check instance type per provider?

	return nil
}
