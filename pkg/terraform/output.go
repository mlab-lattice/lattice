package terraform

import (
	"encoding/json"
)

func Output(workDirectory string, config *Config, outputVars []string) (map[string]string, error) {
	tec, err := NewTerrafromExecContext(workDirectory, nil)
	if err != nil {
		return nil, err
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	err = tec.AddFile("config.tf.json", configBytes)
	if err != nil {
		return nil, err
	}

	result, _, err := tec.Init()
	if err != nil {
		return nil, err
	}

	err = result.Wait()
	if err != nil {
		return nil, err
	}

	return tec.Outputs(outputVars)
}
