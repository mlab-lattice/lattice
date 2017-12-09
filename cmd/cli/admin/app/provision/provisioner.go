package provision

import (
	"fmt"

	kubeprovisioner "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/provisioner"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/lifecycle/provisioner"
)

func getKubernetesProvisioner(provider, systemName string, config *backendConfigKubernetes) (provisioner.Interface, error) {
	switch provider {
	case constants.ProviderLocal:
		lp, err := getLocalProvisioner(config)
		if err != nil {
			return nil, err
		}
		return provisioner.Interface(lp), nil

	case constants.ProviderAWS:
		ap, err := getAWSProvisioner(systemName, config)
		if err != nil {
			return nil, err
		}
		return provisioner.Interface(ap), nil

	default:
		panic(fmt.Sprintf("unsupported provider: %v", provider))
	}
}

func getLocalProvisioner(config *backendConfigKubernetes) (*kubeprovisioner.LocalProvisioner, error) {
	return kubeprovisioner.NewLocalProvisioner(config.DockerAPIVersion, config.LatticeContainerRegistry, config.LatticeContainerRepoPrefix, workingDir+"logs")
}

func getAWSProvisioner(name string, config *backendConfigKubernetes) (*kubeprovisioner.AWSProvisioner, error) {
	awsWorkingDir := workingDir + "/aws/" + name
	awsConfig := config.ProviderConfig.(kubeprovisioner.AWSProvisionerConfig)
	return kubeprovisioner.NewAWSProvisioner(config.LatticeContainerRegistry, config.LatticeContainerRepoPrefix, awsWorkingDir, awsConfig)
}
