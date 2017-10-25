package minikube

import (
	"os/exec"
	"path/filepath"

	executil "github.com/mlab-lattice/core/pkg/util/exec"
)

const (
	binaryName = "minikube"
	startCmd   = "start"
	deleteCmd  = "delete"

	logsDir = "logs"
)

type ExecContext struct {
	*executil.Context
	systemName string
}

func NewMinikubeExecContext(workingDir string, systemName string) (*ExecContext, error) {
	execPath, err := exec.LookPath(binaryName)
	if err != nil {
		return nil, err
	}

	ec, err := executil.NewContext(execPath, workingDir, filepath.Join(workingDir, logsDir))
	if err != nil {
		return nil, err
	}

	mec := &ExecContext{
		Context:    ec,
		systemName: systemName,
	}
	return mec, nil
}

func (mec *ExecContext) Start() (int, string, func() error, error) {
	args := []string{startCmd, "-p", mec.systemName, "--kubernetes-version", "v1.8.0", "--bootstrapper", "kubeadm"}
	return mec.Exec(args...)
}

func (mec *ExecContext) Delete() (int, string, func() error, error) {
	args := []string{deleteCmd, "-p", mec.systemName}
	return mec.Exec(args...)
}
