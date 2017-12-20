package minikube

import (
	"os/exec"
	"strings"

	executil "github.com/mlab-lattice/system/pkg/util/exec"
)

const (
	binaryName = "minikube"
	startCmd   = "start"
	deleteCmd  = "delete"
	ipCmd      = "ip"
)

type ExecContext struct {
	*executil.Context
}

func NewMinikubeExecContext(logPath string) (*ExecContext, error) {
	execPath, err := exec.LookPath(binaryName)
	if err != nil {
		return nil, err
	}

	ec, err := executil.NewContext(execPath, &logPath, nil, nil)
	if err != nil {
		return nil, err
	}

	mec := &ExecContext{
		Context: ec,
	}
	return mec, nil
}

func (mec *ExecContext) Start(name string) (*executil.Result, string, error) {
	// FIXME: make Kubernetes version configurable
	args := []string{startCmd, "-p", name, "--kubernetes-version", "v1.9.0", "--bootstrapper", "kubeadm", "--extra-config=apiserver.v=5", "--extra-config=controller-manager.v=5"}
	return mec.ExecWithLogFile("minikube-"+startCmd, args...)
}

func (mec *ExecContext) Delete(name string) (*executil.Result, string, error) {
	args := []string{deleteCmd, "-p", name}
	return mec.ExecWithLogFile("minikube-"+deleteCmd, args...)
}

func (mec *ExecContext) IP(name string) (string, error) {
	args := []string{ipCmd, "-p", name}
	stdout, _, err := mec.ExecSync(args...)
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(stdout), nil
}
