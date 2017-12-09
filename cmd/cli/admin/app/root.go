package app

import (
	"fmt"
	"os"

	"github.com/mlab-lattice/system/pkg/managerapi/client/admin"
	"github.com/mlab-lattice/system/pkg/managerapi/client/admin/rest"

	"github.com/spf13/cobra"
)

var (
	workingDir  string
	url         string
	adminClient admin.Client
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
	cobra.OnInitialize(initCmd)
	RootCmd.PersistentFlags().StringVar(&workingDir, "working-directory", "/tmp/lattice-system/", "path where subcommands will use as their working directory")
	RootCmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	RootCmd.MarkPersistentFlagRequired("url")
}

func initCmd() {
	err := os.MkdirAll(workingDir, 0770)
	if err != nil {
		panic(fmt.Errorf("unable to create log-path: %v", err))
	}

	adminClient = rest.NewClient(url)
}
