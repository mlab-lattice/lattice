package terraform

import (
	"encoding/json"
)

type Config struct {
	Backend  interface{}
	Provider interface{}
	Modules  map[string]interface{}
	Output   map[string]ConfigOutput
}

type ConfigOutput struct {
	Value string `json:"value"`
}

func (c Config) MarshalJSON() ([]byte, error) {
	jsonMap := map[string]interface{}{
		"provider": c.Provider,
	}

	if c.Modules != nil {
		jsonMap["module"] = c.Modules
	}

	if c.Output != nil {
		jsonMap["output"] = c.Output
	}

	if c.Backend != nil {
		jsonMap["terraform"] = map[string]interface{}{
			"backend": c.Backend,
		}
	}

	return json.Marshal(jsonMap)
}
