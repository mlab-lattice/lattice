package terraform

import (
	"encoding/json"
)

func Destroy(workDirectory string, config *Config) (string, error) {
	tec, err := NewTerrafromExecContext(workDirectory, nil)
	if err != nil {
		return "", err
	}

	if config != nil {
		configBytes, err := json.Marshal(config)
		if err != nil {
			return "", err
		}

		err = tec.AddFile("config.tf.json", configBytes)
		if err != nil {
			return "", err
		}
	}

	result, logfile, err := tec.Init()
	if err != nil {
		return logfile, err
	}

	err = result.Wait()
	if err != nil {
		return logfile, err
	}

	result, logfile, err = tec.Destroy(nil)
	if err != nil {
		return logfile, err
	}

	return logfile, result.Wait()
}
