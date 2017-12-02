package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
)

var systemBuildCmd = &cobra.Command{
	Use:  "system-build",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var systemBuildListCmd = &cobra.Command{
	Use:   "list",
	Short: "list system builds",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		builds, err := namespaceClient.SystemBuilds()
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

var systemBuildGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get system build",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.SystemBuildID(args[0])
		build, err := namespaceClient.SystemBuild(id).Get()
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
	RootCmd.AddCommand(systemBuildCmd)

	systemBuildCmd.AddCommand(systemBuildListCmd)
	systemBuildCmd.AddCommand(systemBuildGetCmd)
}
