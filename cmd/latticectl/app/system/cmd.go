package system

import (
	"log"
	"os"

	"github.com/mlab-lattice/system/cmd/latticectl/app/system/componentbuild"
	"github.com/mlab-lattice/system/cmd/latticectl/app/system/rollout"
	"github.com/mlab-lattice/system/cmd/latticectl/app/system/service"
	"github.com/mlab-lattice/system/cmd/latticectl/app/system/servicebuild"
	"github.com/mlab-lattice/system/cmd/latticectl/app/system/systembuild"
	"github.com/mlab-lattice/system/pkg/cli"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
)

var (
	output        string
	url           string
	clusterClient client.Interface
)

var Cmd = &cobra.Command{
	Use:   "system",
	Short: "commands for managing lattice systems",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list systems",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		systems, err := clusterClient.Systems().List()
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowSystems(systems, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get system",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := types.SystemID(args[0])
		system, err := clusterClient.Systems().Get(id)
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowSystem(system, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

func init() {
	cobra.OnInitialize(initCmd)

	Cmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")

	Cmd.AddCommand(getCmd)
	getCmd.Flags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")

	Cmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")

	Cmd.AddCommand(componentbuild.Cmd)
	Cmd.AddCommand(rollout.Cmd)
	Cmd.AddCommand(service.Cmd)
	Cmd.AddCommand(servicebuild.Cmd)
	Cmd.AddCommand(systembuild.Cmd)
}

func initCmd() {
	clusterClient = rest.NewClient(url)
}
