package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
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

	// Flags().StringArray --provider-var=a,b --provider-var=c results in ["a,b", "c"],
	// whereas Flags().StringSlice --provider-var=a,b --provider-var=c results in ["a", "b", "c"].
	// We don't want this because we want to be able to pass in for example
	// --provider-var=availability-zones=us-east-1a,us-east-1b resulting in ["availability-zones=us-east-1a,us-east-1b"]
	providerVars = provisionSystemCmd.Flags().StringArray("provider-var", nil, "additional variables to pass to the provider")
}
