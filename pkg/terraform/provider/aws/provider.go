package aws

import (
	"encoding/json"
)

type Provider struct {
	Region string
}

// MarshalJSON implements json.Marshaler
func (p Provider) MarshalJSON() ([]byte, error) {
	moduleMap := map[string]interface{}{
		"aws": map[string]interface{}{
			"region": p.Region,
		},
	}
	return json.Marshal(moduleMap)
}
