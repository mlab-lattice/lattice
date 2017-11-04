package app

import (
	"fmt"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"

	sysenvlifecycle "github.com/mlab-lattice/kubernetes-integration/pkg/system-environment/lifecycle"

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

		var provisioner sysenvlifecycle.Provisioner
		switch providerName {
		case coreconstants.ProviderLocal:
			lp, err := sysenvlifecycle.NewLocalProvisioner(devDockerRegistry, logPath)
			if err != nil {
				panic(err)
			}
			provisioner = sysenvlifecycle.Provisioner(lp)
		default:
			panic(fmt.Sprintf("unsupported provider: %v", providerName))
		}

		err := provisioner.Provision(name, url)
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
