package componentbuild

import (
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
)

var (
	follow bool
	output string

	systemIDString string
	url            string
	systemID       types.SystemID
	userClient     client.Interface
	systemClient   client.SystemClient
)

var Cmd = &cobra.Command{
	Use:  "component-build",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list component builds",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		builds, err := systemClient.ComponentBuilds(systemID).List()
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowComponentBuilds(builds, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get component build",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ComponentBuildID(args[0])
		build, err := systemClient.ComponentBuilds(systemID).Get(id)
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowComponentBuild(build, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "get component build logs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ComponentBuildID(args[0])
		logs, err := systemClient.ComponentBuilds(systemID).Logs(id, follow)
		if err != nil {
			log.Fatal(err)
		}
		cli.ShowComponentBuildLog(logs)
		logs.Close()
	},
}

func init() {
	cobra.OnInitialize(initCmd)

	Cmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	Cmd.PersistentFlags().StringVar(&systemIDString, "system", string(constants.SystemIDDefault), "system to use")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "whether or not to follow the logs")
	getCmd.Flags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")
	listCmd.Flags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")
}

func initCmd() {
	systemID = types.SystemID(systemIDString)

	userClient = rest.NewClient(url)
	systemClient = userClient.Systems()
}
