package provision

import (
	"fmt"
	"strings"

	kubeprovisioner "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/provisioner"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/lifecycle/provisioner"
	"github.com/mlab-lattice/system/pkg/util/cli"

	"github.com/spf13/cobra"
)

var (
	workingDir  string
	backend     string
	backendVars []string
)

var Cmd = &cobra.Command{
	Use:   "provision [PROVIDER] [NAME] [URL]",
	Short: "Provision a system",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		provider := args[0]
		name := args[1]
		url := args[2]

		var provisioner provisioner.Interface
		switch backend {
		case constants.BackendTypeKubernetes:
			err := parseBackendKubernetesVars(provider)
			if err != nil {
				panic(fmt.Sprintf("error parsing kubernetes backend vars: %v", err))
			}

			provisioner, err = getKubernetesProvisioner(provider, name)
			if err != nil {
				panic(err)
			}
		default:
			panic(fmt.Sprintf("unsupported backend %v", backend))
		}

		err := provisioner.Provision(name, url)
		if err != nil {
			panic(err)
		}

		addr, err := provisioner.Address(name)
		if err != nil {
			panic(err)
		}

		fmt.Printf("System Environment Manager address:\n%v\n", addr)
	},
}

func init() {
	Cmd.Flags().StringVar(&workingDir, "working-directory", "/tmp/lattice-system/", "path where subcommands will use as their working directory")
	Cmd.Flags().StringVar(&backend, "backend", constants.BackendTypeKubernetes, "lattice backend to use")
	Cmd.Flags().StringArrayVar(&backendVars, "backend-var", nil, "additional variables to pass in to the backend")
}

var backendConfigKubernetes = struct {
	DockerAPIVersion           string
	LatticeContainerRegistry   string
	LatticeContainerRepoPrefix string
	ProviderConfig             interface{}
}{}

func parseBackendKubernetesVars(provider string) error {
	vars := cli.EmbeddedFlag{
		Target: &backendConfigKubernetes,
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
				awsConfig := kubeprovisioner.AWSProvisionerConfig{}
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
		return fmt.Errorf("unsupported provider %v", provider)
	}

	err := vars.Parse(backendVars)
	if err != nil {
		return err
	}

	return nil
}
