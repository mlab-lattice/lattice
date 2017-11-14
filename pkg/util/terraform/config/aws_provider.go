package config

import (
	"encoding/json"
)

type AWSProvider struct {
	Region string
}

// Implement json.Marshaler
func (p AWSProvider) MarshalJSON() ([]byte, error) {
	moduleMap := map[string]interface{}{
		"aws": map[string]interface{}{
			"region": p.Region,
		},
	}
	return json.Marshal(moduleMap)
}
