package servicebuild

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
	systemIDString string
	url            string
	systemID       types.SystemID
	userClient     client.Interface
	systemClient   client.SystemClient
	output         string
)

var Cmd = &cobra.Command{
	Use:  "service-build",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list service builds",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		builds, err := systemClient.ServiceBuilds()
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowServiceBuilds(builds, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get service build",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.ServiceBuildID(args[0])
		build, err := systemClient.ServiceBuild(id).Get()
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowServiceBuild(build, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

func init() {
	cobra.OnInitialize(initCmd)

	Cmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	Cmd.PersistentFlags().StringVar(&systemIDString, "system", string(constants.SystemIDDefault), "system to use")
	Cmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
}

func initCmd() {
	systemID = types.SystemID(systemIDString)

	userClient = rest.NewClient(url)
	systemClient = userClient.System(systemID)
}
