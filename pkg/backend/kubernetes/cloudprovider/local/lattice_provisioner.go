package local

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client/rest"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/minikube"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	clusterNamePrefixMinikube = "lattice-local-"
)

type DefaultLocalLatticeProvisioner struct {
	mec *minikube.ExecContext
}

func NewLatticeProvisioner(workingDir string) (*DefaultLocalLatticeProvisioner, error) {
	mec, err := minikube.NewMinikubeExecContext(fmt.Sprintf("%v/logs", workingDir))
	if err != nil {
		return nil, err
	}

	provisioner := &DefaultLocalLatticeProvisioner{
		mec: mec,
	}
	return provisioner, nil
}

func (p *DefaultLocalLatticeProvisioner) Provision(id v1.LatticeID, containerChannel string, apiAuthKey string) (string, error) {
	prefixedName := clusterNamePrefixMinikube + string(id)
	result, logFilename, err := p.mec.Start(prefixedName)
	if err != nil {
		return "", err
	}

	fmt.Printf("Running minikube start (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*p.mec.LogPath, logFilename))

	err = result.Wait()
	if err != nil {
		return "", err
	}

	address, err := p.address(id)
	if err != nil {
		return "", err
	}

	//err = p.bootstrap(containerChannel, address, apiAuthKey, id)
	//if err != nil {
	//	return "", err
	//}

	fmt.Println("Waiting for API server to be ready...")
	clusterClient := rest.NewUnauthenticatedClient(address)
	err = wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		ok, _ := clusterClient.Health()
		return ok, nil
	})

	if err != nil {
		return "", err
	}

	return address, nil
}

func (p *DefaultLocalLatticeProvisioner) address(id v1.LatticeID) (string, error) {
	prefixedName := clusterNamePrefixMinikube + string(id)
	address, err := p.mec.IP(prefixedName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%v", address), nil
}

func (p *DefaultLocalLatticeProvisioner) Deprovision(name string, force bool) error {
	result, logFilename, err := p.mec.Delete(clusterNamePrefixMinikube + name)
	if err != nil {
		return err
	}

	fmt.Printf("Running minikube delete (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*p.mec.LogPath, logFilename))

	return result.Wait()
}

func getLatticeContainerImage(channel, image string) string {
	return fmt.Sprintf("%v/%v", channel, image)
}
