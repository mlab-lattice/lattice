package app

import (
	"fmt"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"

	sysenvlifecycle "github.com/mlab-lattice/kubernetes-integration/pkg/system-environment/lifecycle"
)

func getProvisioner(provider, systemName string) (sysenvlifecycle.Provisioner, error) {
	switch provider {
	case coreconstants.ProviderLocal:
		lp, err := getLocalProvisioner()
		if err != nil {
			return nil, err
		}
		return sysenvlifecycle.Provisioner(lp), nil

	case coreconstants.ProviderAWS:
		ap, err := getAwsProvisioner(systemName)
		if err != nil {
			return nil, err
		}
		return sysenvlifecycle.Provisioner(ap), nil

	default:
		panic(fmt.Sprintf("unsupported provider: %v", provider))
	}
}

func getLocalProvisioner() (*sysenvlifecycle.LocalProvisioner, error) {
	return sysenvlifecycle.NewLocalProvisioner(devDockerRegistry, workingDir+"logs")
}

func getAwsProvisioner(name string) (*sysenvlifecycle.AWSProvisioner, error) {
	awsWorkingDir := workingDir + "/aws/" + name
	// FIXME: fill in with passed in vars
	awsConfig := sysenvlifecycle.AWSProvisionerConfig{

	}
	return sysenvlifecycle.NewAWSProvisioner(devDockerRegistry, awsWorkingDir, awsConfig)
}
