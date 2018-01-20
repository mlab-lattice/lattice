package provision

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/lifecycle/cluster/provisioner"
	"github.com/mlab-lattice/system/pkg/util/cli"
)

func getKubernetesProvisioner(providerName string) (provisioner.Interface, error) {
	provisionerOptions, err := parseBackendKubernetesVars(providerName)
	if err != nil {
		panic(err)
	}

	return cloudprovider.NewClusterProvisioner(
		provisionerOptions.LatticeContainerRegistry,
		provisionerOptions.LatticeContainerRepoPrefix,
		workDir,
		&provisionerOptions.ProvisionerOptions,
	)
}

type backendConfigKubernetes struct {
	LatticeContainerRegistry   string
	LatticeContainerRepoPrefix string
	ProvisionerOptions         cloudprovider.ClusterProvisionerOptions
}

func parseBackendKubernetesVars(providerName string) (*backendConfigKubernetes, error) {
	config := backendConfigKubernetes{}

	vars := cli.EmbeddedFlag{
		Target: &config,
		Expected: map[string]cli.EmbeddedFlagValue{
			"lattice-container-registry": {
				Default:      "gcr.io/lattice-dev",
				EncodingName: "LatticeContainerRegistry",
			},
			"lattice-container-repo-prefix": {
				Default:      "stable-debug-",
				EncodingName: "LatticeContainerRepoPrefix",
			},
		},
	}

	switch providerName {
	case cloudprovider.Local:
		config.ProvisionerOptions = cloudprovider.ClusterProvisionerOptions{
			Local: &local.ClusterProvisionerOptions{},
		}

	case cloudprovider.AWS:
		vars.Expected["provider-var"] = cli.EmbeddedFlagValue{
			Required: true,
			Array:    true,
			ArrayValueParser: func(values []string) (interface{}, error) {
				awsProvisionerOptions := aws.ClusterProvisionerOptions{}
				providerVars := cli.EmbeddedFlag{
					Target: &awsProvisionerOptions,
					Expected: map[string]cli.EmbeddedFlagValue{
						"terraform-module-path": {
							Default:      "/etc/terraform/modules",
							EncodingName: "TerraformModulePath",
						},
						"terraform-backend-s3-bucket": {
							Required:     true,
							EncodingName: "TerraformBackendS3Bucket",
						},
						"terraform-backend-s3-key": {
							Required:     true,
							EncodingName: "TerraformBackendS3Key",
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

				config.ProvisionerOptions = cloudprovider.ClusterProvisionerOptions{
					AWS: &awsProvisionerOptions,
				}
				return awsProvisionerOptions, nil
			},
		}

	default:
		return nil, fmt.Errorf("unsupported provider %v", providerName)
	}

	err := vars.Parse(backendVars)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
