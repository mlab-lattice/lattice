package app

import (
	"fmt"
	"os"

	"github.com/mlab-lattice/system/cmd/latticectl/app/cluster"
	"github.com/mlab-lattice/system/cmd/latticectl/app/system"

	"github.com/spf13/cobra"
)

// Cmd represents the base command when called without any subcommands
var Cmd = &cobra.Command{
	Use:   "latticectl",
	Short: "Command line utility for interacting with lattice clusters and systems",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := Cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	Cmd.AddCommand(cluster.Cmd)
	Cmd.AddCommand(system.Cmd)
}
