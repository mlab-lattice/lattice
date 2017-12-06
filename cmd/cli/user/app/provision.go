package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	dockerAPIVersion           string
	latticeContainerRegistry   string
	latticeContainerRepoPrefix string

	providerVars *[]string
)

var provisionSystemCmd = &cobra.Command{
	Use:   "provision [PROVIDER] [NAME] [URL]",
	Short: "Provision a system",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		name := args[1]
		url := args[2]

		provisioner, err := getProvisioner(providerName, name, actionProvision, *providerVars)
		if err != nil {
			panic(err)
		}

		err = provisioner.Provision(name, url)
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
	RootCmd.AddCommand(provisionSystemCmd)

	provisionSystemCmd.Flags().StringVar(&dockerAPIVersion, "docker-api-version", "", "version of the docker API used by the docker daemons")
	provisionSystemCmd.Flags().StringVar(&latticeContainerRegistry, "lattice-container-registry", "", "registry which stores the lattice infrastructure containers")
	provisionSystemCmd.Flags().StringVar(&latticeContainerRepoPrefix, "lattice-container-repo-prefix", "", "prefix to append to expected docker image name")

	provisionSystemCmd.MarkFlagRequired("lattice-container-registry")

	// Flags().StringArray --provider-var=a,b --provider-var=c results in ["a,b", "c"],
	// whereas Flags().StringSlice --provider-var=a,b --provider-var=c results in ["a", "b", "c"].
	// We don't want this because we want to be able to pass in for example
	// --provider-var=availability-zones=us-east-1a,us-east-1b resulting in ["availability-zones=us-east-1a,us-east-1b"]
	providerVars = provisionSystemCmd.Flags().StringArray("provider-var", nil, "additional variables to pass to the provider")
}
