package app

import (
	"fmt"

	kubeprovisioner "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/provisioner"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/lifecycle/provisioner"
)

var (
	actionProvision   = "provision"
	actionDeprovision = "deprovision"
)

func getKubernetesProvisioner(provider, systemName, action string, config *backendConfigKubernetes) (provisioner.Interface, error) {
	switch provider {
	case constants.ProviderLocal:
		lp, err := getLocalProvisioner(config)
		if err != nil {
			return nil, err
		}
		return provisioner.Interface(lp), nil

	case constants.ProviderAWS:
		ap, err := getAWSProvisioner(systemName, action, config)
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

func getAWSProvisioner(name, action string, config *backendConfigKubernetes) (*kubeprovisioner.AWSProvisioner, error) {
	awsWorkingDir := workingDir + "/aws/" + name

	var awsConfig kubeprovisioner.AWSProvisionerConfig
	if action == actionProvision {
		awsConfigVars := config.ProviderConfig.(backendConfigKubernetesProviderAWS)
		awsConfig = kubeprovisioner.AWSProvisionerConfig{
			TerraformModulePath: awsConfigVars.ModulePath,

			AccountID:         awsConfigVars.AccountID,
			Region:            awsConfigVars.Region,
			AvailabilityZones: awsConfigVars.AvailabilityZones,
			KeyName:           awsConfigVars.KeyName,

			MasterNodeInstanceType: awsConfigVars.MasterNodeInstanceType,
			MasterNodeAMIID:        awsConfigVars.MasterNodeAMIID,
			BaseNodeAMIID:          awsConfigVars.BaseNodeAMIID,
		}
	}

	return kubeprovisioner.NewAWSProvisioner(config.LatticeContainerRegistry, config.LatticeContainerRepoPrefix, awsWorkingDir, awsConfig)
}
