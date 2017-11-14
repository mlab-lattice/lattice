package config

import (
	"encoding/json"
)

type Module struct {
	Source    string
	Variables map[string]interface{}
}

// Implement json.Marshaler
func (m Module) MarshalJSON() ([]byte, error) {
	moduleMap := map[string]interface{}{}
	if len(m.Variables) > 0 {
		for k, v := range m.Variables {
			moduleMap[k] = v
		}
	}

	moduleMap["source"] = m.Source
	return json.Marshal(moduleMap)
}
