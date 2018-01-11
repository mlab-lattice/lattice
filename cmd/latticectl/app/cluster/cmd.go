package cluster

import (
	"os"

	"github.com/mlab-lattice/system/cmd/latticectl/app/cluster/deprovision"
	"github.com/mlab-lattice/system/cmd/latticectl/app/cluster/kubernetes"
	"github.com/mlab-lattice/system/cmd/latticectl/app/cluster/provision"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cluster",
	Short: "commands for managing a lattice cluster",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

func init() {
	Cmd.AddCommand(deprovision.Cmd)
	Cmd.AddCommand(kubernetes.Cmd)
	Cmd.AddCommand(provision.Cmd)
}
