package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/spf13/cobra"
)

var (
	componentBuildLogsFollow bool
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

var componentBuildGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get component build",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := coretypes.ComponentBuildID(args[0])
		build, err := namespaceClient.ComponentBuild(id).Get()
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

var componentBuildLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "get component build logs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := coretypes.ComponentBuildID(args[0])
		logs, err := namespaceClient.ComponentBuild(id).Logs(componentBuildLogsFollow)
		if err != nil {
			log.Fatal(err)
		}
		defer logs.Close()

		io.Copy(os.Stdout, logs)
	},
}

func init() {
	RootCmd.AddCommand(componentBuildCmd)

	componentBuildCmd.AddCommand(componentBuildListCmd)

	componentBuildCmd.AddCommand(componentBuildGetCmd)

	componentBuildCmd.AddCommand(componentBuildLogsCmd)
	componentBuildLogsCmd.Flags().BoolVar(&componentBuildLogsFollow, "follow", false, "whether or not to follow the logs")

}
