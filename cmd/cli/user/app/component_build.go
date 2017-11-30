package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var componentBuildCmd = &cobra.Command{
	Use:  "component-build",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var componentBuildListCmd = &cobra.Command{
	Use:   "list",
	Short: "list component builds",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		builds, err := namespaceClient.ComponentBuilds()
		if err != nil {
			log.Fatal(err)
		}

		buf, err := json.MarshalIndent(builds, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(buf))
	},
}

func init() {
	RootCmd.AddCommand(componentBuildCmd)

	componentBuildCmd.AddCommand(componentBuildListCmd)
}
