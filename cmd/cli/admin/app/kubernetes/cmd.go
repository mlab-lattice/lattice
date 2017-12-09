package kubernetes

import (
	"os"

	"github.com/mlab-lattice/system/cmd/cli/admin/app/kubernetes/bootstrap"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "commands for managing a Kubernets backend",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

func init() {
	Cmd.AddCommand(bootstrap.Cmd)
}
