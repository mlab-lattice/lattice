package terraform

import (
	"encoding/json"
)

type Config struct {
	Backend  interface{}
	Provider interface{}
	Modules  map[string]interface{}
}

func (c Config) MarshalJSON() ([]byte, error) {
	jsonMap := map[string]interface{}{
		"provider": c.Provider,
		"module":   c.Modules,
	}

	if c.Backend != nil {
		jsonMap["terraform"] = map[string]interface{}{
			"backend": c.Backend,
		}
	}

	return json.Marshal(jsonMap)
}
