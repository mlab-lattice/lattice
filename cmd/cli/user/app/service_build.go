package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
)

var serviceBuildCmd = &cobra.Command{
	Use:  "service-build",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var serviceBuildListCmd = &cobra.Command{
	Use:   "list",
	Short: "list service builds",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		builds, err := namespaceClient.ServiceBuilds()
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

var serviceBuildGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get service build",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ServiceBuildID(args[0])
		build, err := namespaceClient.ServiceBuild(id).Get()
		if err != nil {
			log.Fatal(err)
		}

		buf, err := json.MarshalIndent(build, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(buf))
	},
}

func init() {
	RootCmd.AddCommand(serviceBuildCmd)

	serviceBuildCmd.AddCommand(serviceBuildListCmd)
	serviceBuildCmd.AddCommand(serviceBuildGetCmd)
}
