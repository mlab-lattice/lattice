package app

import (
	"github.com/spf13/cobra"
)

// listBuildsCmd represents the listBuilds command
var deprovisionSystemCmd = &cobra.Command{
	Use:   "deprovision-system [PROVIDER] [NAME]",
	Short: "Deprovision a system",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		name := args[1]

		provisioner, err := getProvisioner(providerName, name)
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
}
