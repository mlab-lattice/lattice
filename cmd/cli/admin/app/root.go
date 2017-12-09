package app

import (
	"fmt"
	"os"

	"github.com/mlab-lattice/system/cmd/cli/admin/app/deprovision"
	"github.com/mlab-lattice/system/cmd/cli/admin/app/kubernetes"
	"github.com/mlab-lattice/system/cmd/cli/admin/app/provision"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "lattice-admin",
	Short: "The lattice-system admin CLI tool",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(deprovision.Cmd)
	RootCmd.AddCommand(kubernetes.Cmd)
	RootCmd.AddCommand(provision.Cmd)
}
