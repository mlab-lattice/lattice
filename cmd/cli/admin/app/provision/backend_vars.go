package provision

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/provisioner"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/util/cli"
)

type backendConfigKubernetes struct {
	DockerAPIVersion           string
	LatticeContainerRegistry   string
	LatticeContainerRepoPrefix string
	ProviderConfig             interface{}
}

func parseBackendKubernetesVars(provider string) (*backendConfigKubernetes, error) {
	config := &backendConfigKubernetes{}
	vars := cli.EmbeddedFlag{
		Target: config,
		Expected: map[string]cli.EmbeddedFlagValue{
			"docker-api-version": {
				EncodingName: "DockerAPIVersion",
			},
			"lattice-container-registry": {
				Required:     true,
				EncodingName: "LatticeContainerRegistry",
			},
			"lattice-container-repo-prefix": {
				EncodingName: "LatticeContainerRepoPrefix",
			},
		},
	}

	switch provider {
	case constants.ProviderLocal:
		// nothing to add
	case constants.ProviderAWS:
		vars.Expected["provider-var"] = cli.EmbeddedFlagValue{
			Required:     true,
			EncodingName: "ProviderConfig",
			Array:        true,
			ArrayValueParser: func(values []string) (interface{}, error) {
				awsConfig := provisioner.AWSProvisionerConfig{}
				providerVars := cli.EmbeddedFlag{
					Target: &awsConfig,
					Expected: map[string]cli.EmbeddedFlagValue{
						"module-path": {
							Required:     true,
							EncodingName: "TerraformModulePath",
						},
						"account-id": {
							Required:     true,
							EncodingName: "AccountID",
						},
						"region": {
							Required:     true,
							EncodingName: "Region",
						},
						"availability-zones": {
							Required:     true,
							EncodingName: "AvailabilityZones",
							ValueParser: func(value string) (interface{}, error) {
								return strings.Split(value, ","), nil
							},
						},
						"key-name": {
							Required:     true,
							EncodingName: "KeyName",
						},
						"master-node-instance-type": {
							Required:     true,
							EncodingName: "MasterNodeInstanceType",
						},
						"master-node-ami-id": {
							Required:     true,
							EncodingName: "MasterNodeAMIID",
						},
						"base-node-ami-id": {
							Required:     true,
							EncodingName: "BaseNodeAMIID",
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
