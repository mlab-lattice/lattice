package terraform

import (
	"encoding/json"
)

func Apply(workDirectory string, config *Config) (string, error) {

	tec, err := NewTerrafromExecContext(workDirectory, nil)
	if err != nil {
		return "", err
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	err = tec.AddFile("config.tf.json", configBytes)
	if err != nil {
		return "", err
	}

	result, logfile, err := tec.Init()
	if err != nil {
		return logfile, err
	}

	err = result.Wait()
	if err != nil {
		return logfile, err
	}

	result, logfile, err = tec.Apply(nil)
	if err != nil {
		return logfile, err
	}

	return logfile, result.Wait()
}
