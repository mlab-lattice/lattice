package deprovision

import (
	"fmt"

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
		"",
		"",
		workDir,
		provisionerOptions,
	)
}

func parseBackendKubernetesVars(providerName string) (*cloudprovider.ClusterProvisionerOptions, error) {
	options := &cloudprovider.ClusterProvisionerOptions{}

	switch providerName {
	case cloudprovider.Local:
		options.Local = &local.ClusterProvisionerOptions{}

	case cloudprovider.AWS:
		vars := cli.EmbeddedFlag{
			Target: &options,
			Expected: map[string]cli.EmbeddedFlagValue{
				"provider-var": {
					Required: true,
					Array:    true,
					ArrayValueParser: func(values []string) (interface{}, error) {
						awsProvisionerOptions := aws.ClusterProvisionerOptions{}
						providerVars := cli.EmbeddedFlag{
							Target: &awsProvisionerOptions,
							Expected: map[string]cli.EmbeddedFlagValue{
								"terraform-backend-s3-bucket": {
									Required:     true,
									EncodingName: "TerraformBackendS3Bucket",
								},
								"terraform-backend-s3-key": {
									Required:     true,
									EncodingName: "TerraformBackendS3Key",
								},
								"cluster-manager-url": {
									EncodingName: "ClusterManagerURL",
								},
							},
						}

						err := providerVars.Parse(values)
						if err != nil {
							return nil, err
						}

						if !force && awsProvisionerOptions.ClusterManagerURL == "" {
							return nil, fmt.Errorf("must specify cluster-manager-url or force")
						}

						options.AWS = &awsProvisionerOptions
						return awsProvisionerOptions, nil
					},
				},
			},
		}

		err := vars.Parse(backendVars)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported provider %v", providerName)
	}

	return options, nil
}
