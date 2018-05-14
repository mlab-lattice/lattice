package terraform

import (
	"encoding/json"
	"os/exec"
	"syscall"
)

type PlanResult int32

const (
	PlanResultEmpty PlanResult = iota
	PlanResultNotEmpty
	PlanResultError
)

func Plan(workDirectory string, config *Config, destroy bool) (PlanResult, string, error) {
	tec, err := NewTerrafromExecContext(workDirectory, nil)
	if err != nil {
		return PlanResultError, "", err
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return PlanResultError, "", err
	}

	err = tec.AddFile("config.tf.json", configBytes)
	if err != nil {
		return PlanResultError, "", err
	}

	result, logfile, err := tec.Init()
	if err != nil {
		return PlanResultError, logfile, err
	}

	err = result.Wait()
	if err != nil {
		return PlanResultError, logfile, err
	}

	result, logfile, err = tec.Plan(nil, destroy)
	if err != nil {
		return PlanResultError, logfile, err
	}

	err = result.Wait()
	if err == nil {
		return PlanResultEmpty, logfile, nil
	}

	// try to get the exit code (https://stackoverflow.com/questions/10385551/get-exit-code-go)
	if exitError, ok := err.(*exec.ExitError); ok {
		ws := exitError.Sys().(syscall.WaitStatus)
		exitCode := ws.ExitStatus()
		switch exitCode {
		case 0:
			// shouldn't happen
			return PlanResultEmpty, logfile, nil

		case 2:
			return PlanResultNotEmpty, logfile, nil

		default:
			return PlanResultError, logfile, err
		}
	}

	// This will happen (in OSX) if `name` is not available in $PATH,
	// in this situation, exit code could not be gotten
	return PlanResultError, logfile, err
}
