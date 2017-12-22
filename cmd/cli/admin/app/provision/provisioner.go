package provision

import (
	"fmt"

	kubeprovisioner "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/provisioner"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/lifecycle/cluster/provisioner"
)

func getKubernetesProvisioner(provider, systemName string) (provisioner.Interface, error) {
	switch provider {
	case constants.ProviderLocal:
		lp, err := getLocalProvisioner()
		if err != nil {
			return nil, err
		}
		return provisioner.Interface(lp), nil

	case constants.ProviderAWS:
		ap, err := getAWSProvisioner(systemName)
		if err != nil {
			return nil, err
		}
		return provisioner.Interface(ap), nil

	default:
		panic(fmt.Sprintf("unsupported provider: %v", provider))
	}
}

func getLocalProvisioner() (*kubeprovisioner.LocalProvisioner, error) {
	return kubeprovisioner.NewLocalProvisioner(backendConfigKubernetes.DockerAPIVersion, backendConfigKubernetes.LatticeContainerRegistry, backendConfigKubernetes.LatticeContainerRepoPrefix, workingDir+"logs")
}

func getAWSProvisioner(name string) (*kubeprovisioner.AWSProvisioner, error) {
	awsWorkingDir := workingDir + "/aws/" + name

	if backendConfigKubernetes.ProviderConfig.AWS == nil {
		return nil, fmt.Errorf("AWSConfig not set")
	}

	awsConfig := *backendConfigKubernetes.ProviderConfig.AWS
	return kubeprovisioner.NewAWSProvisioner(backendConfigKubernetes.LatticeContainerRegistry, backendConfigKubernetes.LatticeContainerRepoPrefix, awsWorkingDir, awsConfig)
}
