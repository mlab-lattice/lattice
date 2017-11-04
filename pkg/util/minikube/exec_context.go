package minikube

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"

	executil "github.com/mlab-lattice/core/pkg/util/exec"
	"strings"
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

	ec, err := executil.NewContext(execPath, logPath, nil)
	if err != nil {
		return nil, err
	}

	mec := &ExecContext{
		Context: ec,
	}
	return mec, nil
}

func (mec *ExecContext) Start(name string) (int, string, func() error, error) {
	args := []string{startCmd, "-p", name, "--kubernetes-version", "v1.8.0", "--bootstrapper", "kubeadm"}
	return mec.Exec("minikube-"+startCmd, args...)
}

func (mec *ExecContext) Delete(name string) (int, string, func() error, error) {
	args := []string{deleteCmd, "-p", name}
	return mec.Exec("minikube-"+deleteCmd, args...)
}

func (mec *ExecContext) IP(name string) (string, error) {
	args := []string{ipCmd, "-p", name}
	_, logFilename, waitFunc, err := mec.Exec("minikube-"+ipCmd, args...)
	if err != nil {
		return "", err
	}

	err = waitFunc()
	if err != nil {
		return "", err
	}

	ipBytes, err := ioutil.ReadFile(filepath.Join(mec.LogPath, logFilename))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(ipBytes)), nil
}
