package app

import (
	"github.com/spf13/cobra"
)

var (
	deprovisionVars *[]string
)

var deprovisionSystemCmd = &cobra.Command{
	Use:   "deprovision-system [PROVIDER] [NAME]",
	Short: "Deprovision a system",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		name := args[1]

		provisioner, err := getProvisioner(providerName, name, *deprovisionVars)
		if err != nil {
			panic(err)
		}

		err = provisioner.Deprovision(name)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(deprovisionSystemCmd)

	// Flags().StringArray --provider-var=a,b --provider-var=c results in ["a,b", "c"],
	// whereas Flags().StringSlice --provider-var=a,b --provider-var=c results in ["a", "b", "c"].
	// We don't want this because we want to be able to pass in for example
	// --provider-var=availability-zones=us-east-1a,us-east-1b resulting in ["availability-zones=us-east-1a,us-east-1b"]
	deprovisionVars = deprovisionSystemCmd.Flags().StringArray("provider-var", nil, "additional variables to pass to the provider")
}
