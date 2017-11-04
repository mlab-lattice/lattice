package app

import (
	"fmt"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"

	sysenvlifecycle "github.com/mlab-lattice/kubernetes-integration/pkg/system-environment/lifecycle"

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

		err := provisioner.Deprovision(name)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(deprovisionSystemCmd)
}
