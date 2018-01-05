package terraform

import (
	"encoding/json"
)

func Apply(workDirectory string, config *Config) error {
	tec, err := NewTerrafromExecContext(workDirectory, nil)
	if err != nil {
		return err
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = tec.AddFile("config.tf", configBytes)
	if err != nil {
		return err
	}

	result, _, err := tec.Init()
	if err != nil {
		return err
	}

	err = result.Wait()
	if err != nil {
		return err
	}

	result, _, err = tec.Apply(nil)
	if err != nil {
		return err
	}

	return result.Wait()
}
