package system

import (
	"os"

	"github.com/mlab-lattice/system/cmd/latticectl/app/system/componentbuild"
	"github.com/mlab-lattice/system/cmd/latticectl/app/system/servicebuild"
	"github.com/mlab-lattice/system/cmd/latticectl/app/system/systembuild"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "system",
	Short: "commands for managing lattice systems",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

func init() {
	Cmd.AddCommand(componentbuild.Cmd)
	Cmd.AddCommand(servicebuild.Cmd)
	Cmd.AddCommand(systembuild.Cmd)
}
