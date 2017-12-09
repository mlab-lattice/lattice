package app

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/util/cli"
)

type backendConfigKubernetes struct {
	DockerAPIVersion           string      `json:"docker-api-version"`
	LatticeContainerRegistry   string      `json:"lattice-container-registry"`
	LatticeContainerRepoPrefix string      `json:"lattice-container-repo-prefix"`
	ProviderConfig             interface{} `json:"provider-var"`
}

type backendConfigKubernetesProviderAWS struct {
	ModulePath             string   `json:"module-path"`
	AccountID              string   `json:"account-id"`
	Region                 string   `json:"region"`
	AvailabilityZones      []string `json:"availability-zones"`
	KeyName                string   `json:"key-name"`
	MasterNodeInstanceType string   `json:"master-node-instance-type"`
	MasterNodeAMIID        string   `json:"master-node-ami-id"`
	BaseNodeAMIID          string   `json:"base-node-ami-id"`
}

func parseBackendKubernetesVars(provider string) (*backendConfigKubernetes, error) {
	config := &backendConfigKubernetes{}
	vars := cli.EmbeddedFlag{
		Target: config,
		Expected: map[string]cli.EmbeddedFlagValue{
			"docker-api-version": {},
			"lattice-container-registry": {
				Required: true,
			},
			"lattice-container-repo-prefix": {},
		},
	}

	switch provider {
	case constants.ProviderLocal:
		// nothing to add
	case constants.ProviderAWS:
		vars.Expected["provider-var"] = cli.EmbeddedFlagValue{
			Required: true,
			Array:    true,
			ArrayValueParser: func(values []string) (interface{}, error) {
				awsConfig := backendConfigKubernetesProviderAWS{}
				providerVars := cli.EmbeddedFlag{
					Target: &awsConfig,
					Expected: map[string]cli.EmbeddedFlagValue{
						"module-path": {
							Required: true,
						},
						"account-id": {
							Required: true,
						},
						"region": {
							Required: true,
						},
						"availability-zones": {
							Required: true,
							ValueParser: func(value string) (interface{}, error) {
								return strings.Split(value, ","), nil
							},
						},
						"key-name": {
							Required: true,
						},
						"master-node-instance-type": {
							Required: true,
						},
						"master-node-ami-id": {
							Required: true,
						},
						"base-node-ami-id": {
							Required: true,
						},
					},
				}

				err := providerVars.Parse(values)
				if err != nil {
					return nil, err
				}
				return awsConfig, nil
			},
		}
	default:
		return nil, fmt.Errorf("unsupported provider %v", provider)
	}

	err := vars.Parse(backendVars)
	if err != nil {
		return nil, err
	}

	return config, nil
}
