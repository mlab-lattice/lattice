package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	devDockerRegistry  = "gcr.io/lattice-dev"
)

var (
	logPath string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "lattice-system",
	Short: "The lattice-system CLI tool",
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
	RootCmd.PersistentFlags().StringVar(&logPath, "log-path", "/tmp/lattice-system/logs", "path where logs from subcommands will be stored")
}

func initCmd() {
	err := os.MkdirAll(logPath, 0770)
	if err != nil {
		panic(fmt.Errorf("unable to create log-path: %v", err))
	}
}
