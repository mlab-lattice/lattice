package app

import (
	"fmt"
	"strings"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"

	sysenvlifecycle "github.com/mlab-lattice/kubernetes-integration/pkg/system-environment/lifecycle"
)

func getProvisioner(provider, systemName string, providerVars []string) (sysenvlifecycle.Provisioner, error) {
	switch provider {
	case coreconstants.ProviderLocal:
		lp, err := getLocalProvisioner()
		if err != nil {
			return nil, err
		}
		return sysenvlifecycle.Provisioner(lp), nil

	case coreconstants.ProviderAWS:
		ap, err := getAwsProvisioner(systemName, providerVars)
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

func getAwsProvisioner(name string, providerVars []string) (*sysenvlifecycle.AWSProvisioner, error) {
	awsWorkingDir := workingDir + "/aws/" + name

	// TODO: find a better way to do the parsing of the provider variables
	expectedVars := map[string]interface{}{
		"module-path":               nil,
		"account-id":                nil,
		"region":                    nil,
		"availability-zones":        nil,
		"key-name":                  nil,
		"master-node-instance-type": nil,
		"master-node-ami-id":        nil,
		"base-node-ami-id":          nil,
	}

	for _, providerVar := range providerVars {
		split := strings.Split(providerVar, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid provider variable " + providerVar)
		}

		key := split[0]
		var value interface{} = split[1]

		existingVal, ok := expectedVars[key]
		if !ok {
			return nil, fmt.Errorf("unexpected provider variable " + key)
		}
		if existingVal != nil {
			return nil, fmt.Errorf("provider variable " + key + " set multiple times")
		}

		if key == "availability-zones" {
			value = strings.Split(value.(string), ",")
		}

		expectedVars[key] = value
	}

	for k, v := range expectedVars {
		if v == nil {
			return nil, fmt.Errorf("missing required provider variable " + k)
		}
	}

	awsConfig := sysenvlifecycle.AWSProvisionerConfig{
		ModulePath:        expectedVars["module-path"].(string),
		AccountId:         expectedVars["account-id"].(string),
		Region:            expectedVars["region"].(string),
		AvailabilityZones: expectedVars["availability-zones"].([]string),
		KeyName:           expectedVars["key-name"].(string),

		MasterNodeInstanceType: expectedVars["master-node-instance-type"].(string),
		MasterNodeAMIId:        expectedVars["master-node-ami-id"].(string),
		BaseNodeAmiId:          expectedVars["base-node-ami-id"].(string),
	}

	return sysenvlifecycle.NewAWSProvisioner(devDockerRegistry, awsWorkingDir, awsConfig)
}
