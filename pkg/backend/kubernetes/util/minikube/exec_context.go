package minikube

import (
	"os/exec"
	"strings"

	executil "github.com/mlab-lattice/lattice/pkg/util/exec"
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
	// FIXME: add back profile name when supported: https://github.com/kubernetes/minikube/issues/2574
	//args := []string{startCmd, "-p", name, "--kubernetes-version", "v1.9.3", "--bootstrapper", "kubeadm", "--feature-gates=CustomPodDNS=true"}
	args := []string{startCmd, "--kubernetes-version", "v1.10.0", "--bootstrapper", "kubeadm", "--memory", "4096"}
	return mec.ExecWithLogFile("minikube-"+startCmd, args...)
}

func (mec *ExecContext) Delete(name string) (*executil.Result, string, error) {
	// FIXME: add back profile name when supported: https://github.com/kubernetes/minikube/issues/2574
	args := []string{deleteCmd}
	//args := []string{deleteCmd, "-p", name}
	return mec.ExecWithLogFile("minikube-"+deleteCmd, args...)
}

func (mec *ExecContext) IP(name string) (string, error) {
	// FIXME: add back profile name when supported: https://github.com/kubernetes/minikube/issues/2574
	args := []string{ipCmd}
	//args := []string{ipCmd, "-p", name}
	stdout, _, err := mec.ExecSync(args...)
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(stdout), nil
}
