package terraform

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	executil "github.com/mlab-lattice/core/pkg/util/exec"
)

const (
	binaryName = "terraform"
	applyCmd   = "apply"
	destroyCmd = "destroy"
	initCmd    = "init"
	outputCmd  = "output"

	logsDir = "logs"
)

type ExecContext struct {
	*executil.Context
}

func NewTerrafromExecContext(workingDir string, envVars map[string]string) (*ExecContext, error) {
	execPath, err := exec.LookPath(binaryName)
	if err != nil {
		return nil, err
	}

	logPath := filepath.Join(workingDir, logsDir)
	ec, err := executil.NewContext(execPath, &logPath, &workingDir, envVars)
	if err != nil {
		return nil, err
	}

	tec := &ExecContext{
		Context: ec,
	}
	return tec, nil
}

func (tec *ExecContext) Apply(vars map[string]string) (*executil.Result, string, error) {
	args := []string{applyCmd, "-input=false", "-auto-approve=true", "-no-color"}

	for k, v := range vars {
		args = append(args, fmt.Sprintf("-var='%s=%s'", k, v))
	}

	return tec.ExecWithLogFile("terraform-"+applyCmd, args...)
}

func (tec *ExecContext) Destroy(vars map[string]string) (*executil.Result, string, error) {
	args := []string{destroyCmd, "-force", "-no-color"}

	for k, v := range vars {
		args = append(args, fmt.Sprintf("-var='%s=%s'", k, v))
	}

	return tec.ExecWithLogFile("terraform-"+destroyCmd, args...)
}

func (tec *ExecContext) Init() (*executil.Result, string, error) {
	args := []string{initCmd, "-force-copy"}
	return tec.ExecWithLogFile("terraform-"+initCmd, args...)
}

func (tec ExecContext) Output(outputVar string) (string, error) {
	args := []string{outputCmd, outputVar}
	stdout, _, err := tec.ExecSync(args...)
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(stdout), nil
}
