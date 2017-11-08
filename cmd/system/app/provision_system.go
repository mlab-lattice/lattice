package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listBuildsCmd represents the listBuilds command
var provisionSystemCmd = &cobra.Command{
	Use:   "provision-system [PROVIDER] [NAME] [URL]",
	Short: "Provision a system",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		name := args[1]
		url := args[2]

		provisioner, err := getProvisioner(providerName, name)
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

		fmt.Printf("SystemManager address:\n%v\n", addr)
	},
}

func init() {
	RootCmd.AddCommand(provisionSystemCmd)
}
