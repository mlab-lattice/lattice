package app

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/constants"
	kubelifecycle "github.com/mlab-lattice/system/pkg/kubernetes/lifecycle"
	"github.com/mlab-lattice/system/pkg/lifecycle"
)

var (
	actionProvision   = "provision"
	actionDeprovision = "deprovision"
)

func getKubernetesProvisioner(provider, systemName, action string, config *backendConfigKubernetes) (lifecycle.Provisioner, error) {
	switch provider {
	case constants.ProviderLocal:
		lp, err := getLocalProvisioner(config)
		if err != nil {
			return nil, err
		}
		return lifecycle.Provisioner(lp), nil

	case constants.ProviderAWS:
		ap, err := getAWSProvisioner(systemName, action, config)
		if err != nil {
			return nil, err
		}
		return lifecycle.Provisioner(ap), nil

	default:
		panic(fmt.Sprintf("unsupported provider: %v", provider))
	}
}

func getLocalProvisioner(config *backendConfigKubernetes) (*kubelifecycle.LocalProvisioner, error) {
	return kubelifecycle.NewLocalProvisioner(config.DockerAPIVersion, config.LatticeContainerRegistry, config.LatticeContainerRepoPrefix, workingDir+"logs")
}

func getAWSProvisioner(name, action string, config *backendConfigKubernetes) (*kubelifecycle.AWSProvisioner, error) {
	awsWorkingDir := workingDir + "/aws/" + name

	var awsConfig kubelifecycle.AWSProvisionerConfig
	if action == actionProvision {
		awsConfigVars := config.ProviderConfig.(backendConfigKubernetesProviderAWS)
		awsConfig = kubelifecycle.AWSProvisionerConfig{
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

	return kubelifecycle.NewAWSProvisioner(config.LatticeContainerRegistry, config.LatticeContainerRepoPrefix, awsWorkingDir, awsConfig)
}
